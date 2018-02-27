// Copyright (c) 2017 The Alvalor Authors
//
// This file is part of Alvalor.
//
// Alvalor is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Alvalor is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package codec

import (
	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
	capnp "zombiezen.com/go/capnproto2"
)

func transactionToMessage(entity *types.Transaction) (*capnp.Message, error) {
	msg, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize message")
	}
	z, err := NewRootZ(seg)
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize wrapper")
	}
	transaction, err := z.NewTransaction()
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize transaction")
	}
	transfers, err := transaction.NewTransfers(int32(len(entity.Transfers)))
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize transfer list")
	}
	for i, item := range entity.Transfers {
		transfer, err := NewTransfer(seg)
		if err != nil {
			return nil, errors.Wrap(err, "could not initialize transfer")
		}
		err = transfer.SetFrom(item.From)
		if err != nil {
			return nil, errors.Wrap(err, "could not set from")
		}
		err = transfer.SetTo(item.To)
		if err != nil {
			return nil, errors.Wrap(err, "could not set to")
		}
		transfer.SetAmount(item.Amount)
		err = transfers.Set(i, transfer)
		if err != nil {
			return nil, errors.Wrap(err, "could not set transfer")
		}
	}
	fees, err := transaction.NewFees(int32(len(entity.Fees)))
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize fee list")
	}
	for i, item := range entity.Fees {
		var fee Fee
		fee, err = NewFee(seg)
		if err != nil {
			return nil, errors.Wrap(err, "could not initialize fee")
		}
		fee.SetFrom(item.From)
		fee.SetAmount(item.Amount)
		err = fees.Set(i, fee)
		if err != nil {
			return nil, errors.Wrap(err, "could not set fee")
		}
	}
	err = transaction.SetData(entity.Data)
	if err != nil {
		return nil, errors.Wrap(err, "could not set transaction data")
	}
	sigs, err := transaction.NewSignatures(int32(len(entity.Signatures)))
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize signature list")
	}
	for i, sig := range entity.Signatures {
		err = sigs.Set(i, sig)
		if err != nil {
			return nil, errors.Wrap(err, "could not set signature")
		}
	}
	return msg, nil
}

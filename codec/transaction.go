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
	"github.com/pkg/errors"
	capnp "zombiezen.com/go/capnproto2"

	"github.com/alvalor/alvalor-go/types"
)

type initTransaction func() (Transaction, error)

func rootTransaction(z Z) initTransaction {
	return z.NewTransaction
}

func childTransaction(seg *capnp.Segment) initTransaction {
	return func() (Transaction, error) {
		transaction, err := NewTransaction(seg)
		return transaction, err
	}
}

func encodeTransaction(seg *capnp.Segment, init initTransaction, e *types.Transaction) (Transaction, error) {
	transaction, err := init()
	if err != nil {
		return Transaction{}, errors.Wrap(err, "could not initialize transaction")
	}
	transfers, err := transaction.NewTransfers(int32(len(e.Transfers)))
	if err != nil {
		return Transaction{}, errors.Wrap(err, "could not initialize transfer list")
	}
	for i, t := range e.Transfers {
		var transfer Transfer
		transfer, err = encodeTransfer(seg, childTransfer(seg), &t)
		if err != nil {
			return Transaction{}, errors.Wrap(err, "could not encode transfer")
		}
		err = transfers.Set(i, transfer)
		if err != nil {
			return Transaction{}, errors.Wrap(err, "could not set transfer")
		}
	}
	fees, err := transaction.NewFees(int32(len(e.Fees)))
	if err != nil {
		return Transaction{}, errors.Wrap(err, "could not initialize fee list")
	}
	for i, f := range e.Fees {
		var fee Fee
		fee, err = encodeFee(seg, childFee(seg), &f)
		if err != nil {
			return Transaction{}, errors.Wrap(err, "could not encode fee")
		}
		err = fees.Set(i, fee)
		if err != nil {
			return Transaction{}, errors.Wrap(err, "could not set fee")
		}
	}
	err = transaction.SetData(e.Data)
	if err != nil {
		return Transaction{}, errors.Wrap(err, "could not set data")
	}
	sigs, err := transaction.NewSignatures(int32(len(e.Signatures)))
	if err != nil {
		return Transaction{}, errors.Wrap(err, "could not initialize signature list")
	}
	for i, sig := range e.Signatures {
		err = sigs.Set(i, sig)
		if err != nil {
			return Transaction{}, errors.Wrap(err, "could not set signature")
		}
	}
	return transaction, nil
}

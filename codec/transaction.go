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

func createRootTransaction(z Z) initTransaction {
	return z.NewTransaction
}

func createChildTransaction(seg *capnp.Segment) initTransaction {
	return func() (Transaction, error) {
		transaction, err := NewTransaction(seg)
		return transaction, err
	}
}

func readRootTransaction(z Z) initTransaction {
	return z.Transaction
}

func readChildTransaction(transaction Transaction) initTransaction {
	return func() (Transaction, error) {
		return transaction, nil
	}
}

func encodeTransaction(seg *capnp.Segment, init initTransaction, e *types.Transaction) (Transaction, error) {
	transaction, err := init()
	if err != nil {
		return Transaction{}, errors.Wrap(err, "could not create transaction")
	}
	transfers, err := transaction.NewTransfers(int32(len(e.Transfers)))
	if err != nil {
		return Transaction{}, errors.Wrap(err, "could not create transfer list")
	}
	for i, t := range e.Transfers {
		var transfer Transfer
		transfer, err = encodeTransfer(seg, createChildTransfer(seg), t)
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
		return Transaction{}, errors.Wrap(err, "could not create fee list")
	}
	for i, f := range e.Fees {
		var fee Fee
		fee, err = encodeFee(seg, createChildFee(seg), f)
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
	transaction.SetNonce(e.Nonce)
	sigs, err := transaction.NewSignatures(int32(len(e.Signatures)))
	if err != nil {
		return Transaction{}, errors.Wrap(err, "could not create signature list")
	}
	for i, sig := range e.Signatures {
		err = sigs.Set(i, sig)
		if err != nil {
			return Transaction{}, errors.Wrap(err, "could not set signature")
		}
	}
	return transaction, nil
}

func decodeTransaction(read initTransaction) (*types.Transaction, error) {
	transaction, err := read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read transaction")
	}
	transfers, err := transaction.Transfers()
	if err != nil {
		return nil, errors.Wrap(err, "could not read transfer list")
	}
	fees, err := transaction.Fees()
	if err != nil {
		return nil, errors.Wrap(err, "could not read fee list")
	}
	data, err := transaction.Data()
	if err != nil {
		return nil, errors.Wrap(err, "could not read data")
	}
	signatures, err := transaction.Signatures()
	if err != nil {
		return nil, errors.Wrap(err, "could not read signature list")
	}
	e := &types.Transaction{
		Transfers:  make([]*types.Transfer, 0, transfers.Len()),
		Fees:       make([]*types.Fee, 0, fees.Len()),
		Data:       data,
		Nonce:      transaction.Nonce(),
		Signatures: make([][]byte, 0, signatures.Len()),
	}
	for i := 0; i < transfers.Len(); i++ {
		transfer := transfers.At(i)
		t, err := decodeTransfer(readChildTransfer(transfer))
		if err != nil {
			return nil, errors.Wrap(err, "could not decode transfer")
		}
		e.Transfers = append(e.Transfers, t)
	}
	for i := 0; i < fees.Len(); i++ {
		fee := fees.At(i)
		f, err := decodeFee(readChildFee(fee))
		if err != nil {
			return nil, errors.Wrap(err, "could not decode fee")
		}
		e.Fees = append(e.Fees, f)
	}
	for i := 0; i < signatures.Len(); i++ {
		signature, err := signatures.At(i)
		if err != nil {
			return nil, errors.Wrap(err, "could not get signature")
		}
		e.Signatures = append(e.Signatures, signature)
	}
	return e, nil
}

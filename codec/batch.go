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

	"github.com/alvalor/alvalor-go/node"
	"github.com/alvalor/alvalor-go/types"
)

type initBatch func() (Batch, error)

func createRootBatch(z Z) initBatch {
	return z.NewBatch
}

func readRootBatch(z Z) initBatch {
	return z.Batch
}

func encodeBatch(seg *capnp.Segment, create initBatch, e *node.Batch) (Batch, error) {
	batch, err := create()
	if err != nil {
		return Batch{}, errors.Wrap(err, "could not create batch")
	}
	transactions, err := batch.NewTransactions(int32(len(e.Transactions)))
	if err != nil {
		return Batch{}, errors.Wrap(err, "could not create transaction list")
	}
	for i, t := range e.Transactions {
		var transaction Transaction
		transaction, err = encodeTransaction(seg, createChildTransaction(seg), t)
		if err != nil {
			return Batch{}, errors.Wrap(err, "could not encode transaction")
		}
		err = transactions.Set(i, transaction)
		if err != nil {
			return Batch{}, errors.Wrap(err, "could not set transaction")
		}
	}
	return batch, nil
}

func decodeBatch(read initBatch) (*node.Batch, error) {
	batch, err := read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read batch")
	}
	transactions, err := batch.Transactions()
	if err != nil {
		return nil, errors.Wrap(err, "could not read transaction list")
	}
	e := &node.Batch{
		Transactions: make([]*types.Transaction, 0, transactions.Len()),
	}
	for i := 0; i < transactions.Len(); i++ {
		transaction := transactions.At(i)
		t, err := decodeTransaction(readChildTransaction(transaction))
		if err != nil {
			return nil, errors.Wrap(err, "could not decode transation")
		}
		e.Transactions = append(e.Transactions, t)
	}
	return e, nil
}

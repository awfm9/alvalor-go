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
)

type initBatch func() (Batch, error)

func rootBatch(z Z) initBatch {
	return z.NewBatch
}

func encodeBatch(seg *capnp.Segment, init initBatch, e *node.Batch) (Batch, error) {
	batch, err := init()
	if err != nil {
		return Batch{}, errors.Wrap(err, "could not initialize batch")
	}
	transactions, err := batch.NewTransactions(int32(len(e.Transactions)))
	if err != nil {
		return Batch{}, errors.Wrap(err, "could not initialize transaction list")
	}
	for i, t := range e.Transactions {
		var transaction Transaction
		transaction, err = encodeTransaction(seg, childTransaction(seg), t)
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

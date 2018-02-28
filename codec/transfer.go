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

type initTransfer func() (Transfer, error)

func createChildTransfer(seg *capnp.Segment) initTransfer {
	return func() (Transfer, error) {
		transfer, err := NewTransfer(seg)
		return transfer, err
	}
}

func encodeTransfer(seg *capnp.Segment, create initTransfer, e *types.Transfer) (Transfer, error) {
	transfer, err := create()
	if err != nil {
		return Transfer{}, errors.Wrap(err, "could not create transfer")
	}
	err = transfer.SetFrom(e.From)
	if err != nil {
		return Transfer{}, errors.Wrap(err, "could not set from")
	}
	err = transfer.SetTo(e.To)
	if err != nil {
		return Transfer{}, errors.Wrap(err, "could not set to")
	}
	transfer.SetAmount(e.Amount)
	return transfer, nil
}

func readChildTransfer(transfer Transfer) initTransfer {
	return func() (Transfer, error) {
		return transfer, nil
	}
}

func decodeTransfer(read initTransfer) (*types.Transfer, error) {
	transfer, err := read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read transfer")
	}
	from, err := transfer.From()
	if err != nil {
		return nil, errors.Wrap(err, "could not get from")
	}
	to, err := transfer.To()
	if err != nil {
		return nil, errors.Wrap(err, "could not get to")
	}
	amount := transfer.Amount()
	e := &types.Transfer{
		From:   from,
		To:     to,
		Amount: amount,
	}
	return e, nil
}

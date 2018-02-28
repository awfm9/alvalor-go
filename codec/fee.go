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

type initFee func() (Fee, error)

func createChildFee(seg *capnp.Segment) initFee {
	return func() (Fee, error) {
		fee, err := NewFee(seg)
		return fee, err
	}
}

func encodeFee(seg *capnp.Segment, create initFee, e *types.Fee) (Fee, error) {
	fee, err := create()
	if err != nil {
		return Fee{}, errors.Wrap(err, "could not create fee")
	}
	err = fee.SetFrom(e.From)
	if err != nil {
		return Fee{}, errors.Wrap(err, "could not set from")
	}
	fee.SetAmount(e.Amount)
	return fee, nil
}

func readChildFee(fee Fee) initFee {
	return func() (Fee, error) {
		return fee, nil
	}
}

func decodeFee(read initFee) (*types.Fee, error) {
	fee, err := read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read fee")
	}
	from, err := fee.From()
	if err != nil {
		return nil, errors.Wrap(err, "could not get from")
	}
	amount := fee.Amount()
	f := &types.Fee{
		From:   from,
		Amount: amount,
	}
	return f, nil
}

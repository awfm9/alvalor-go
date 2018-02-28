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

	"github.com/alvalor/alvalor-go/network"
)

type initPong func() (Pong, error)

func createRootPong(z Z) initPong {
	return z.NewPong
}

func readRootPong(z Z) initPong {
	return z.Pong
}

func encodePong(seg *capnp.Segment, create initPong, e *network.Pong) (Pong, error) {
	pong, err := create()
	if err != nil {
		return Pong{}, errors.Wrap(err, "could not initialize pong")
	}
	pong.SetNonce(e.Nonce)
	return pong, nil
}

func decodePong(read initPong) (*network.Pong, error) {
	pong, err := read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read pong")
	}
	e := &network.Pong{
		Nonce: pong.Nonce(),
	}
	return e, nil
}

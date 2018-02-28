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

type initPing func() (Ping, error)

func createRootPing(z Z) initPing {
	return z.NewPing
}

func readRootPing(z Z) initPing {
	return z.Ping
}

func encodePing(seg *capnp.Segment, create initPing, e *network.Ping) (Ping, error) {
	ping, err := create()
	if err != nil {
		return Ping{}, errors.Wrap(err, "could not create ping")
	}
	ping.SetNonce(e.Nonce)
	return ping, nil
}

func decodePing(read initPing) (*network.Ping, error) {
	ping, err := read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read ping")
	}
	e := &network.Ping{
		Nonce: ping.Nonce(),
	}
	return e, nil
}

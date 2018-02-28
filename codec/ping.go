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

func rootPing(z Z) initPing {
	return z.NewPing
}

func childPing(seg *capnp.Segment) initPing {
	return func() (Ping, error) {
		ping, err := NewPing(seg)
		return ping, err
	}
}

func encodePing(seg *capnp.Segment, init initPing, e *network.Ping) (Ping, error) {
	ping, err := init()
	if err != nil {
		return Ping{}, errors.Wrap(err, "could not initialize ping")
	}
	ping.SetNonce(e.Nonce)
	return ping, nil
}

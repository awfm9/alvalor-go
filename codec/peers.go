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

type initPeers func() (Peers, error)

func createRootPeers(z Z) initPeers {
	return z.NewPeers
}

func readRootPeers(z Z) initPeers {
	return z.Peers
}

func encodePeers(seg *capnp.Segment, create initPeers, e *network.Peers) (Peers, error) {
	peers, err := create()
	if err != nil {
		return Peers{}, errors.Wrap(err, "could not create peers")
	}
	addrs, err := peers.NewAddresses(int32(len(e.Addresses)))
	if err != nil {
		return Peers{}, errors.Wrap(err, "could not create address list")
	}
	for i, address := range e.Addresses {
		err = addrs.Set(i, address)
		if err != nil {
			return Peers{}, errors.Wrap(err, "could not set address")
		}
	}
	return peers, nil
}

func decodePeers(read initPeers) (*network.Peers, error) {
	peers, err := read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read peers")
	}
	addresses, err := peers.Addresses()
	if err != nil {
		return nil, errors.Wrap(err, "could not read address list")
	}
	e := &network.Peers{
		Addresses: make([]string, 0, addresses.Len()),
	}
	for i := 0; i < addresses.Len(); i++ {
		address, err := addresses.At(i)
		if err != nil {
			return nil, errors.Wrap(err, "could not get address")
		}
		e.Addresses = append(e.Addresses, address)
	}
	return e, nil
}

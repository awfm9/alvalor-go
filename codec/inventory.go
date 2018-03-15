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

type initInventory func() (Inventory, error)

func createRootInventory(z Z) initInventory {
	return z.NewInventory
}

func readRootInventory(z Z) initInventory {
	return z.Inventory
}

func encodeInventory(seg *capnp.Segment, create initInventory, e *node.Inventory) (Inventory, error) {
	inventory, err := create()
	if err != nil {
		return Inventory{}, errors.Wrap(err, "could not create inventory")
	}
	hashes, err := inventory.NewHashes(int32(len(e.Hashes)))
	if err != nil {
		return Inventory{}, errors.Wrap(err, "could not create hash list")
	}
	for i, hash := range e.Hashes {
		err = hashes.Set(i, hash[:])
		if err != nil {
			return Inventory{}, errors.Wrap(err, "could not set hash")
		}
	}
	return inventory, nil
}

func decodeInventory(read initInventory) (*node.Inventory, error) {
	inventory, err := read()
	if err != nil {
		return nil, errors.Wrap(err, "could not read inventory")
	}
	hashes, err := inventory.Hashes()
	if err != nil {
		return nil, errors.Wrap(err, "could not read hash list")
	}
	e := &node.Inventory{
		Hashes: make([]types.Hash, hashes.Len()),
	}
	for i := 0; i < hashes.Len(); i++ {
		hash, err := hashes.At(i)
		if err != nil {
			return nil, errors.Wrap(err, "could not get hash")
		}
		copy(e.Hashes[i][:], hash)
	}
	return e, nil
}

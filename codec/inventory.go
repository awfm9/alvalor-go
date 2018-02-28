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

type initInventory func() (Inventory, error)

func rootInventory(z Z) initInventory {
	return z.NewInventory
}

func encodeInventory(seg *capnp.Segment, init initInventory, e *node.Inventory) (Inventory, error) {
	inventory, err := init()
	if err != nil {
		return Inventory{}, errors.Wrap(err, "could not initialize inventory")
	}
	ids, err := inventory.NewIds(int32(len(e.IDs)))
	if err != nil {
		return Inventory{}, errors.Wrap(err, "could not initialize ID list")
	}
	for i, id := range e.IDs {
		err = ids.Set(i, id)
		if err != nil {
			return Inventory{}, errors.Wrap(err, "could not set ID")
		}
	}
	return inventory, nil
}

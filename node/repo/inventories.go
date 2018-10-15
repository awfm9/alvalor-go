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

package repo

import (
	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/types"
)

// Inventories is a simple implementation of the inventory store.
type Inventories struct {
	inventories map[types.Hash]*types.Inventory
}

// NewInventories creates a new store for block inventories.
func NewInventories() *Inventories {
	ir := &Inventories{
		inventories: make(map[types.Hash]*types.Inventory),
	}
	// TODO: load all known inventories from disk
	return ir
}

// Add stores a new inventory.
func (ir *Inventories) Add(inv *types.Inventory) error {
	_, ok := ir.inventories[inv.Hash]
	if ok {
		return errors.Wrap(ErrAlreadyExists, "header already known")
	}
	ir.inventories[inv.Hash] = inv
	return nil
}

// Has checks if a given inventory is known.
func (ir *Inventories) Has(hash types.Hash) bool {
	_, ok := ir.inventories[hash]
	return ok
}

// Get retrieves the inventory with the given block hash.
func (ir *Inventories) Get(hash types.Hash) (*types.Inventory, error) {
	inv, ok := ir.inventories[hash]
	if !ok {
		return nil, errors.Wrap(ErrNotFound, "inventory not found")
	}
	return inv, nil
}

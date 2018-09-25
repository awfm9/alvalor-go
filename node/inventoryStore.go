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

package node

import (
	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/types"
)

// inventoryStore represents an interface to block inventory storage.
type inventoryStore interface {
	Inventory(hash types.Hash) (*Inventory, error)
	AddInventory(hash types.Hash, hashes []types.Hash) error
}

// inventoryStoreS is a simple implementation of the inventory store.
type inventoryStoreS struct {
	inventories map[types.Hash]*Inventory
}

// AddInventory stores a new inventory.
func (is *inventoryStoreS) AddInvestory(inv *Inventory) error {
	_, ok := is.inventories[inv.Hash]
	if ok {
		return errors.New("inventory already known")
	}
	is.inventories[inv.Hash] = inv
	return nil
}

// Inventory retrieves the inventory with the given block hash.
func (is *inventoryStoreS) Inventory(hash types.Hash) (*Inventory, error) {
	inv, ok := is.inventories[hash]
	if !ok {
		return nil, errors.Wrap(errNotFound, "inventory not found")
	}
	return inv, nil
}

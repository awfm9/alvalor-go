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

package blocks

import (
	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
)

// Assembler organizes the block downloads.
type Assembler struct {
	pending      map[types.Hash]struct{}
	templates    map[types.Hash]map[types.Hash]bool
	inventories  Inventories
	transactions Transactions
	downloads    Downloads
}

// Assemble starts the assembly of the block with the given hash.
func (as *Assembler) Assemble(hash types.Hash) error {

	// check if we are already downloading this block
	_, ok := as.pending[hash]
	if ok {
		return errors.Wrap(ErrExist, "block assembly already requested")
	}

	// check if we already have the template
	_, ok = as.templates[hash]
	if ok {
		return nil
	}

	// check if we are already downloading the inventory
	ok = as.downloads.HasInv(hash)
	if ok {
		return nil
	}

	// start the inventory download
	err := as.downloads.StartInv(hash)
	if err != nil {
		return errors.Wrap(err, "inventory download already requested")
	}

	// mark the block assembly as pending
	as.pending[hash] = struct{}{}

	return nil
}

// Suspend suspends the assembly of the block with the given hash.
func (as *Assembler) Suspend(hash types.Hash) error {

	// TODO: implement

	return nil
}

// Inventory notifies the block assembler that an inventory was received.
func (as *Assembler) Inventory(hash types.Hash) error {

	// check if we are actually waiting for the inventory
	_, ok := as.pending[hash]
	if !ok {
		return nil
	}

	// check if we already have the template in place
	_, ok = as.templates[hash]
	if ok {
		return nil
	}

	// retrieve the inventory
	inv, err := as.inventories.Get(hash)
	if err != nil {
		return errors.Wrap(err, "could not get inventory to create template")
	}

	// create and save the download template
	template := make(map[types.Hash]bool)
	for _, hash := range inv.Hashes {
		ok := as.transactions.Has(hash)
		if ok {
			template[hash] = true
		} else {
			template[hash] = false
		}
	}

	// start the transaction downloads that are not pending
	for hash, ok := range template {
		if ok {
			continue
		}
		// TODO: what if we already have all of them? can it happen?
		ok = as.downloads.HasTx(hash)
		if ok {
			continue
		}
		err := as.downloads.StartTx(hash)
		if err != nil {
			return errors.Wrap(err, "could not start transaction download for block assembly")
		}
	}

	// save the template for the block
	as.templates[hash] = template

	return nil
}

// Transaction notifies the block downloader when a transaction is received.
func (as *Assembler) Transaction(hash *types.Hash) error {

	// TODO: register and assemble block when it's the last one required
	return nil
}

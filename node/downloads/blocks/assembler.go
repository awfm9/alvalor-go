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
	mapping      map[types.Hash]types.Hash
	headers      Headers
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
	for _, txHash := range inv.Hashes {
		ok := as.transactions.Has(txHash)
		if ok {
			template[txHash] = true
		} else {
			as.mapping[txHash] = hash
			template[txHash] = false
		}
	}

	// TODO: move following to another function

	// start the transaction downloads that are not pending
	for txHash, ok := range template {
		if ok {
			continue
		}
		// TODO: what if we already have all of them? can it happen?
		ok = as.downloads.HasTx(txHash)
		if ok {
			continue
		}
		err := as.downloads.StartTx(txHash)
		if err != nil {
			return errors.Wrapf(err, "could not start transaction download for block assembly: %v", txHash)
		}
	}

	// save the template for the block
	as.templates[hash] = template

	return nil
}

// Transaction notifies the block downloader when a transaction is received.
func (as *Assembler) Transaction(hash types.Hash) error {

	// check if we are waiting for the given transaction
	blkHash, ok := as.mapping[hash]
	if !ok {
		return nil
	}

	// retrieve the block template
	template, ok := as.templates[blkHash]
	if !ok {
		return errors.New("block template missing for transaction")
	}

	// set the given transaction to received
	template[hash] = true

	// if still transactions missing, do nothing
	for _, ok := range template {
		if !ok {
			return nil
		}
	}

	// TODO: move following to another function

	// retrieve the header
	header, err := as.headers.Get(blkHash)
	if err != nil {
		return errors.Wrap(err, "could not retrieve header to finalize block")
	}

	// retrieve the inventory
	inv, err := as.inventories.Get(blkHash)
	if err != nil {
		return errors.Wrap(err, "could not retrieve inventory to finalize block")
	}

	// build block
	block := types.Block{
		Header: header,
	}
	for _, txHash := range inv.Hashes {
		tx, err := as.transactions.Get(txHash)
		if err != nil {
			return errors.Wrapf(err, "could not retrieve transaction to finalize block: %v", txHash)
		}
		block.Transactions = append(block.Transactions, tx)
	}

	// validate block
	// TODO: validation logic

	return nil
}

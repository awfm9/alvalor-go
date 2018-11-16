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

package orchestration

import (
	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
)

// Manager organizes the block downloads.
type Manager struct {
	pending      map[types.Hash]struct{}
	templates    map[types.Hash]map[types.Hash]bool
	mapping      map[types.Hash]types.Hash
	download     Download
	assembly     Assembly
	inventories  Inventories
	transactions Transactions
}

// Collect starts collecting all entities required to assemble a block.
func (om *Manager) Collect(hash types.Hash) error {

	// check if we are already downloading this block
	_, ok := om.pending[hash]
	if ok {
		return errors.Wrap(ErrExist, "block collection already started")
	}

	// check if we already have the template
	_, ok = om.templates[hash]
	if ok {
		return errors.Wrap(ErrExist, "block template already exists")
	}

	// check if we are already downloading the inventory
	ok = om.download.HasInv(hash)
	if ok {
		return errors.Wrap(ErrExist, "inventory download already started")
	}

	// start the inventory download
	err := om.download.StartInv(hash)
	if err != nil {
		return errors.Wrap(err, "could not start inventory download")
	}

	// mark the block assembly as pending
	om.pending[hash] = struct{}{}

	return nil
}

// Suspend suspends the assembly of the block with the given hash.
func (om *Manager) Suspend(hash types.Hash) error {

	// check if we are currently collecting for the given hash
	_, ok := om.pending[hash]
	if !ok {
		return errors.Wrap(ErrNotExist, "block collection not started")
	}

	// check if we have an inventory download pending and cancel it
	delete(om.pending, hash)
	ok = om.download.HasInv(hash)
	if ok {
		err := om.download.CancelInv(hash)
		if err != nil {
			return errors.Wrap(err, "could not cancel inventory download")
		}
	}

	// check if we have a block template
	template, ok := om.templates[hash]
	if !ok {
		return nil
	}

	// cancel all transactions downloads and delete the mapping
	delete(om.templates, hash)
	for txHash := range template {
		ok = om.download.HasTx(txHash)
		if !ok {
			continue
		}
		err := om.download.CancelTx(txHash)
		if err != nil {
			return errors.Wrap(err, "could not cancel transaction download")
		}
	}

	return nil
}

// Inventory notifies the block assembler that an inventory was received.
func (om *Manager) Inventory(hash types.Hash) error {

	// check if we are actually waiting for the inventory
	_, ok := om.pending[hash]
	if !ok {
		return errors.Wrap(ErrNotExist, "no pending block assembly")
	}

	// check if we already have the template
	_, ok = om.templates[hash]
	if ok {
		return errors.Wrap(ErrExist, "block template already exists")
	}

	// retrieve the inventory
	inv, err := om.inventories.Get(hash)
	if err != nil {
		return errors.Wrap(err, "could not get inventory to create template")
	}

	// create and save the download template
	template := make(map[types.Hash]bool)
	for _, txHash := range inv.Hashes {
		ok := om.transactions.Has(txHash)
		if ok {
			template[txHash] = true
		} else {
			om.mapping[txHash] = hash
			template[txHash] = false
		}
	}

	// TODO: move following to another function

	// start the transaction downloads that are not pending
	for txHash, ok := range template {
		if ok {
			continue
		}
		err := om.download.StartTx(txHash)
		if err != nil {
			return errors.Wrapf(err, "could not start transaction download (%v)", txHash)
		}
	}

	// save the template for the block
	om.templates[hash] = template

	return nil
}

// Transaction notifies the block downloader when a transaction is received.
func (om *Manager) Transaction(hash types.Hash) error {

	// check if we are waiting for the given transaction
	blkHash, ok := om.mapping[hash]
	if !ok {
		return errors.Wrap(ErrNotExist, "not waiting for transaction download")
	}

	// retrieve the block template
	template, ok := om.templates[blkHash]
	if !ok {
		return errors.Wrap(ErrNotExist, "block template missing for transaction")
	}

	// set the given transaction to received
	template[hash] = true

	// if still transactions missing, do nothing
	for _, ok := range template {
		if !ok {
			return nil
		}
	}

	// at this point, start block validation
	err := om.assembly.Validate(hash)
	if err != nil {
		return errors.Wrap(err, "could not validate block")
	}

	return nil
}

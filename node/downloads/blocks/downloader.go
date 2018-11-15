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

// Downloader organizes the block downloads.
type Downloader struct {
	pending     map[types.Hash]struct{}
	inventories Inventories
	downloads   Downloads
}

// Start starts the download of entities needed for a block.
func (do *Downloader) Start(hash types.Hash) error {

	// check if we are already downloading this block
	_, ok := do.pending[hash]
	if ok {
		return errors.Wrap(ErrExist, "download for block already running")
	}

	// check if we already have the inventory
	ok = do.inventories.Has(hash)
	if ok {
		return do.Inventory(hash)
	}

	// check if we already download the inventory
	ok = do.downloads.HasInv(hash)
	if ok {
		return errors.New("inventory download for block already running")
	}

	// start the inventory download
	err := do.downloads.StartInv(hash)
	if err != nil {
		return errors.Wrap(err, "could not start inventory download for block")
	}

	return nil
}

// Cancel stop the download of the entities needed for a block.
func (do *Downloader) Cancel(hash types.Hash) error {

	// check if we are currently downloading this block
	_, ok := do.pending[hash]
	if !ok {
		return errors.Wrap(ErrNotExist, "download for block not running")
	}

	// TODO: cancel stuff

	return nil
}

// Inventory notifies the block downloader when an inventory is received.
func (do *Downloader) Inventory(hash types.Hash) error {

	// check if we are actually waiting for the inventory
	_, ok := do.pending[hash]
	if !ok {
		return errors.New("received inventory we are not waiting for")
	}

	// retrieve the inventory
	inv, err := do.inventories.Get(hash)
	if err != nil {
		return errors.Wrap(err, "could not get inventory for block download")
	}

	// start the transaction downloads
	for _, hash := range inv.Hashes {
		ok := do.downloads.HasTx(hash)
		if ok {
			continue
		}
		err := do.downloads.StartTx(hash)
		if err != nil {
			return errors.Wrap(err, "could not start transaction download for block")
		}
	}

	return nil
}

// Transaction notifies the block downloader when a transaction is received.
func (do *Downloader) Transaction(hash *types.Hash) error {

	// TODO: register and assemble block when it's the last one required
	return nil
}

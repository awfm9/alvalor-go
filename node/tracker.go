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

type tracker interface {
	Follow(path []types.Hash) error
	Signal(hash types.Hash) error
}

type trackerS struct {
	inventories inventoryStore
	downloads   downloader
	current     map[types.Hash]bool
	running     map[types.Hash]<-chan struct{}
}

func newTracker() *trackerS {
	return &trackerS{
		running: make(map[types.Hash]<-chan struct{}),
	}
}

// Follow sets a new path through the header tree to follow and complete.
func (tr *trackerS) Follow(path []types.Hash) error {

	// NOTE: There might be a big overlap of transactions between two competing
	// paths. Canceling all downloads and restarting them when a majority of them
	// could be in common is not efficient. We should find the difference
	// between the inventories, as far as available.

	// first we collect all transactions on the new path that we know we will
	// have to synchronize
	newTxs := make(map[types.Hash]struct{})
	lookup := make(map[types.Hash]struct{})

	// for each hash on the new path, starting at the oldest header hash
	for _, hash := range path {

		// create a lookup map for the new path hashes
		lookup[hash] = struct{}{}

		// if we are already synchronizing this hash, we can ignore it
		_, ok := tr.current[hash]
		if ok {
			continue
		}

		// now, we add it to the map as "false" indicating we are waiting for the
		// related inventory to start transaction downloads
		tr.current[hash] = false

		// if we do not find the inventory, we start the download
		inv, err := tr.inventories.Inventory(hash)
		if errors.Cause(err) == errNotFound {
			inErr := tr.downloads.StartInventory(inv.Hash)
			if inErr != nil {
				return errors.Wrap(inErr, "could not start inventory download")
			}
			continue
		}
		if err != nil {
			return errors.Wrap(err, "could not retrieve new inventory")
		}

		// at this point, we already have the inventory and will start the related
		// transaction downloads, so change the map entry to true
		tr.current[hash] = true

		// let's now add the hashes from the inventory of the block header to the
		// set of transactions we want to download on this path
		for _, txHash := range inv.Hashes {
			newTxs[txHash] = struct{}{}
		}
	}

	// second, we collect all transactions that we were already synchronizing on
	// the old path
	oldTxs := make(map[types.Hash]struct{})

	// for each hash on the currently synchronizing path
	for hash := range tr.current {

		// first, we check for each header if they are on the new path as well and,
		// if they are, we ignore them
		_, ok := lookup[hash]
		if ok {
			continue
		}

		// if they are not on the new path, we remove the header from our currently
		// synchronizing inventories/headers
		delete(tr.current, hash)

		// we then retrieve the inventory to cancel any transaction downloads that
		// might be in progress and not needed for the new headers
		inv, err := tr.inventories.Inventory(hash)
		if errors.Cause(err) == errNotFound {
			// NOTE: no need to cancel related transactions
			continue
		}
		if err != nil {
			return errors.Wrap(err, "could not retrieve old inventory")
		}

		// we also make sure to cancel the download of the inventory for each of
		// the headers on this path, as they are no longer needed
		err = tr.downloads.CancelInventory(hash)
		if err != nil && errors.Cause(err) != errNotFound {
			return errors.Wrap(err, "could not cancel old inventory download")
		}

		// for those headers where we already have the inventory, we create a set
		// of old transactions that are already synchronizing
		for _, txHash := range inv.Hashes {
			oldTxs[txHash] = struct{}{}
		}
	}

	// we can start the download of all new transactions that were not being
	// synchronized yet (they are in newTxs and not in oldTxs)
	for txHash := range newTxs {
		_, ok := oldTxs[txHash]
		if !ok {
			err := tr.downloads.StartTransaction(txHash)
			if err != nil {
				return errors.Wrap(err, "could not start new transaction download")
			}
		}
	}

	// we can cancel the download of all old transactions that are no longer
	// needed (they are in oldTxs and not in newTxs)
	for txHash := range oldTxs {
		_, ok := newTxs[txHash]
		if !ok {
			err := tr.downloads.CancelTransaction(txHash)
			if err != nil {
				return errors.Wrap(err, "could not cancel old transaction download")
			}
		}
	}

	return nil
}

// Signal notifies the tracker than a new inventory has become available and
// related transaction downloads should be started, if pending.
func (tr *trackerS) Signal(hash types.Hash) error {
	already, ok := tr.current[hash]
	if !ok || already {
		return nil
	}
	inv, err := tr.inventories.Inventory(hash)
	if err != nil {
		return errors.Wrap(err, "could not retrieve inventory for signal")
	}
	for _, hash := range inv.Hashes {
		err := tr.downloads.StartTransaction(hash)
		if err != nil {
			return errors.Wrap(err, "could not start transaction download on signal")
		}
	}
	return nil
}

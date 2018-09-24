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

import "github.com/alvalor/alvalor-go/types"

type downloader interface {
	Follow(hash []types.Hash)
	Complete(hash types.Hash)
	Abort(hash types.Hash)
}

type simpleDownloader struct {
	current map[types.Hash]struct{}
	running map[types.Hash]<-chan struct{}
}

func newSimpleDownloader() *simpleDownloader {
	return &simpleDownloader{
		running: make(map[types.Hash]<-chan struct{}),
	}
}

// Follow sets a new path through the header tree to follow and complete.
func (sd *simpleDownloader) Follow(path []types.Hash) {

	// NOTE: There might be a big overlap of transactions between two competing
	// paths. Canceling all downloads and restarting them when a majority of them
	// could be in common is not efficient. We should find the difference
	// between the inventories, as far as available.

	// for each new hash on the path, we start the download of transactions
	newTxs := make(map[types.Hash]struct{})
	lookup := make(map[types.Hash]struct{})
	for _, hash := range path {
		lookup[hash] = struct{}{}
		_, ok := sd.current[hash]
		if ok {
			continue
		}
		sd.current[hash] = struct{}{}
		// TODO: retrieve inventory for header
		var inv []types.Hash
		if !ok {
			// TODO: request inventory for header
			continue
		}
		for _, txHash := range inv {
			newTxs[txHash] = struct{}{}
		}
	}

	// for each hash on the old path that's not on the new one, we cancel it
	oldTxs := make(map[types.Hash]struct{})
	for hash := range sd.current {
		_, ok := lookup[hash]
		if ok {
			continue
		}
		delete(sd.current, hash)
		// TODO: retrieve inventory for header
		var inv []types.Hash
		for _, txHash := range inv {
			oldTxs[txHash] = struct{}{}
		}
	}

	// cancel each old transaction download that is not in the new set
	for txHash := range oldTxs {
		_, ok := newTxs[txHash]
		if !ok {
			// TODO: cancel transaction download
		}
	}

	// start each new transaction download that is not in the old set
	for txHash := range newTxs {
		_, ok := oldTxs[txHash]
		if !ok {
			// TODO: start transaction download
		}
	}
}

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
	"math"

	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
)

// downloader manages downloading of entities by keeping track of pending
// downloads and load balancing across available peers.
type downloader interface {
	StartInventory(hash types.Hash) error
	CancelInventory(hash types.Hash) error
	StartTransaction(hash types.Hash) error
	CancelTransaction(hash types.Hash) error
}

// downloaderS implement a simple download manager.
type downloaderS struct {
	inventories  map[types.Hash]string
	transactions map[types.Hash]string
	peers        peerManager
	net          Network
}

// newDownloader creates a new simple download manager.
func newDownloader() *downloaderS {
	return &downloaderS{}
}

// StartInventory starts the download of a block inventory.
func (do *downloaderS) StartInventory(hash types.Hash) error {

	// if we are already downloading the inventory, skip
	_, ok := do.inventories[hash]
	if ok {
		return nil
	}

	// get the peer with the lowest amount of pending downloads
	addresses := do.peers.Actives()
	var target string
	best := uint(math.MaxUint32)
	for _, address := range addresses {
		pending, err := do.peers.Pending(address)
		if err != nil {
			continue
		}
		if pending >= best {
			continue
		}
		best = pending
		target = address
	}

	// send the request to the given peer
	msg := &Confirm{Hash: hash}
	err := do.net.Send(target, msg)
	if err != nil {
		return errors.Wrap(err, "could not send inventory request")
	}

	// TODO: implement peer state for downloads

	// TODO: implement timeout mechanism

	return nil
}

// CancelInventory cancels the download of a block inventory.
func (do *downloaderS) CancelInventory(hash types.Hash) error {

	// TODO: disable the timeout mechanism

	return nil
}

// StartTransaction starts the download of a transaction.
func (do *downloaderS) StartTransaction(hash types.Hash) error {

	// if we are already downloading the transaction, skip
	_, ok := do.transactions[hash]
	if ok {
		return nil
	}

	return nil
}

// CancelTransaction cancels the download of a transaction.
func (do *downloaderS) CancelTransaction(hash types.Hash) error {

	return nil
}

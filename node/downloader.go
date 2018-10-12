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
	"sync"
	"time"

	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
)

// downloader manages downloading of entities by keeping track of pending
// downloads and load balancing across available peers.
type downloader interface {
	Start(hash types.Hash) error
	Cancel(hash types.Hash) error
}

// downloaderS implement a simple download manager.
type downloaderS struct {
	sync.Mutex
	inventories  map[types.Hash]string
	transactions map[types.Hash]string
	timeouts     map[types.Hash]chan<- struct{}
	peers        Peers
	net          Network
}

// newDownloader creates a new simple download manager.
func newDownloader() *downloaderS {
	return &downloaderS{}
}

// Start starts the download of a block inventory.
func (do *downloaderS) Start(hash types.Hash) error {
	do.Lock()
	defer do.Unlock()

	// if we are already downloading the inventory, skip
	_, ok := do.inventories[hash]
	if ok {
		return nil
	}

	// get all active peers that have the desired inventory
	addresses := do.peers.Find(peerIsActive(true), peerHasEntity(true, hash))
	if len(addresses) == 0 {
		return errors.New("no active peers with inventory available")
	}

	// select the available peer with the least amount of pending downloads
	var target string
	best := uint(math.MaxUint32)
	for _, address := range addresses {
		pending, err := do.peers.NumPending(address)
		if err != nil {
			continue
		}
		if pending >= best {
			continue
		}
		best = pending
		target = address
	}

	// send the request to the given peer and mark inventory as requested
	msg := &Request{Hash: hash}
	err := do.net.Send(target, msg)
	if err != nil {
		return errors.Wrap(err, "could not send inventory request")
	}
	do.peers.Requested(target, hash)

	// start a timeout timer to retry the download and save the cancel signal
	cancel := timeout(4*time.Second, func() { do.Start(hash) })
	do.timeouts[hash] = cancel

	return nil
}

// CancelInventory cancels the download of a block inventory.
func (do *downloaderS) Cancel(hash types.Hash) error {
	do.Lock()
	defer do.Unlock()

	cancel, ok := do.timeouts[hash]
	if !ok {
		return errors.Wrap(errNotFound, "could not find download for hash")
	}
	close(cancel)

	return nil
}

func timeout(duration time.Duration, retry func()) chan struct{} {
	cancel := make(chan struct{})
	go func() {
		select {
		case <-time.After(duration):
			retry()
		case <-cancel:
			// nothing
		}
	}()
	return cancel
}

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

package download

import (
	"math"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/node/message"
	"github.com/alvalor/alvalor-go/node/peer"
	"github.com/alvalor/alvalor-go/types"
)

// Manager implement a simple download manager.
type Manager struct {
	sync.Mutex
	peers        peer.State
	net          Network
	inventories  map[types.Hash]string
	transactions map[types.Hash]string
	timeouts     map[types.Hash]chan<- struct{}
}

// Start starts the download of a block inventory.
func (mgr *Manager) Start(hash types.Hash) error {
	mgr.Lock()
	defer mgr.Unlock()

	// if we are already downloading the inventory, skip
	_, ok := mgr.inventories[hash]
	if ok {
		return nil
	}

	// get all active peers that have the desired inventory
	addresses := mgr.peers.Addresses(peer.IsActive(true), peer.HasEntity(true, hash))
	if len(addresses) == 0 {
		return errors.New("no active peers with inventory available")
	}

	// select the available peer with the least amount of pending downloads
	var target string
	best := uint(math.MaxUint32)
	for _, address := range addresses {
		pending, err := mgr.peers.Pending(address)
		if err != nil {
			continue
		}
		if uint(len(pending)) <= best {
			continue
		}
		best = uint(len(pending))
		target = address
	}

	// send the request to the given peer and mark inventory as requested
	msg := &message.Request{Hash: hash}
	err := mgr.net.Send(target, msg)
	if err != nil {
		return errors.Wrap(err, "could not send inventory request")
	}
	mgr.peers.Requested(target, hash)

	// start a timeout timer to retry the download and save the cancel signal
	cancel := timeout(4*time.Second, func() { mgr.Start(hash) })
	mgr.timeouts[hash] = cancel

	return nil
}

// Cancel cancels the download of a block inventory.
func (mgr *Manager) Cancel(hash types.Hash) error {
	mgr.Lock()
	defer mgr.Unlock()

	cancel, ok := mgr.timeouts[hash]
	if !ok {
		return errors.New("could not find download for hash")
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

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

	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/node/message"
	"github.com/alvalor/alvalor-go/node/peer"
	"github.com/alvalor/alvalor-go/types"
)

// Manager implement a simple download manager.
type Manager struct {
	sync.Mutex
	net     Network
	peers   Peers
	pending map[types.Hash]string
}

// NewManager creates a new download manager with initialized maps and injected
// dependencies.
func NewManager(net Network, peers Peers) *Manager {
	return &Manager{
		net:     net,
		peers:   peers,
		pending: make(map[types.Hash]string),
	}
}

// Start starts the download of a block inventory.
func (mgr *Manager) Start(hash types.Hash) error {
	mgr.Lock()
	defer mgr.Unlock()

	// if we are already downloading the entity, skip
	_, ok := mgr.pending[hash]
	if ok {
		return errors.New("download already requested")
	}

	// get all active peers that have the desired entity
	// TODO: we should make the distinction between: has, might, doesn't
	addresses := mgr.peers.Addresses(peer.IsActive(true), peer.HasEntity(true, hash))
	if len(addresses) == 0 {
		return errors.New("no active peers with entity available")
	}

	// create a lookup map of all peers who have the entity
	count := make(map[string]uint)
	for _, address := range addresses {
		count[address] = 0
	}

	// check how many downloads are pending for each fitting peer
	for _, address := range mgr.pending {
		_, ok := count[address]
		if !ok {
			continue
		}
		count[address]++
	}

	// select the available peer with the least amount of pending download
	var address string
	best := uint(math.MaxUint32)
	for _, candidate := range addresses {
		if count[candidate] >= best {
			continue
		}
		best = count[candidate]
		address = candidate
	}

	// send the request to the given peer
	msg := &message.Request{Hash: hash}
	err := mgr.net.Send(address, msg)
	if err != nil {
		return errors.Wrap(err, "could not send inventory request")
	}

	// mark the request as pending for this peer
	mgr.pending[hash] = address

	// TODO: add timeout & retry functionality

	return nil
}

// Cancel cancels the download of a block inventory.
func (mgr *Manager) Cancel(hash types.Hash) error {
	mgr.Lock()
	defer mgr.Unlock()

	// find which peer is currently pending for this download
	_, ok := mgr.pending[hash]
	if !ok {
		return errors.New("could not find download for hash")
	}

	// TODO: cancel timeout and abort download

	// remove the pending entry
	delete(mgr.pending, hash)

	return nil
}

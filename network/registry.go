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

package network

import "sync"

// registry represents a simple map of peers that is safe for concurrent access.
type registry struct {
	mutex sync.RWMutex
	peers map[string]*peer
}

// slice returns a copied slice of all registered peers.
func (reg *registry) slice() []*peer {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	peers := make([]*peer, 0, len(reg.peers))
	for _, peer := range reg.peers {
		peers = append(peers, peer)
	}
	return peers
}

// has returns true if we know a peer with the given address.
func (reg *registry) has(addr string) bool {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	_, ok := reg.peers[addr]
	return ok
}

// remove will remove the peer with the given address from the registry.
func (reg *registry) remove(addr string) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	delete(reg.peers, addr)
}

// add will add the given peer with the given address to the registry.
func (reg *registry) add(addr string, peer *peer) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	reg.peers[addr] = peer
}

// count will return the number of peers currently in the registry.
func (reg *registry) count() int {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	return len(reg.peers)
}

// get will return the peer with the given address.
func (reg *registry) get(addr string) *peer {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	return reg.peers[addr]
}

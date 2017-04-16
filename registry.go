// Copyright (c) 2017 The Veltor Authors
//
// This file is part of Veltor.
//
// Veltor is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Veltor is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Veltor.  If not, see <http://www.gnu.org/licenses/>.

package network

import "sync"

// registry struct.
type registry struct {
	mutex sync.RWMutex
	peers map[string]*peer
}

// slice method.
func (reg *registry) slice() []*peer {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	peers := make([]*peer, 0, len(reg.peers))
	for _, peer := range reg.peers {
		peers = append(peers, peer)
	}
	return peers
}

// has method.
func (reg *registry) has(addr string) bool {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	_, ok := reg.peers[addr]
	return ok
}

// remove method.
func (reg *registry) remove(addr string) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	delete(reg.peers, addr)
}

// add method.
func (reg *registry) add(addr string, peer *peer) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	reg.peers[addr] = peer
}

// count method.
func (reg *registry) count() int {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	return len(reg.peers)
}

// get method.
func (reg *registry) get(addr string) *peer {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	return reg.peers[addr]
}

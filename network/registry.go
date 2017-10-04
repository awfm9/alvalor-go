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

import (
	"net"
	"sync"
	"time"

	"github.com/pierrec/lz4"
)

// Registry represents the registry used to manage peers.
type Registry struct {
	mutex sync.RWMutex
	peers map[string]*peer
}

// slice returns a copied slice of all registered peers.
func (reg *Registry) slice() []*peer {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	peers := make([]*peer, 0, len(reg.peers))
	for _, peer := range reg.peers {
		peers = append(peers, peer)
	}
	return peers
}

// has returns true if we know a peer with the given address.
func (reg *Registry) has(addr string) bool {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	_, ok := reg.peers[addr]
	return ok
}

// remove will remove the peer with the given address from the registry.
func (reg *Registry) remove(addr string) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	delete(reg.peers, addr)
}

// add will add the given peer with the given address to the registry.
func (reg *Registry) add(address string, conn net.Conn, nonce []byte) {
	r := lz4.NewReader(conn)
	w := lz4.NewWriter(conn)
	outgoing := make(chan interface{}, 16)
	incoming := make(chan interface{}, 16)
	p := &peer{
		conn:      conn,
		address:   address,
		nonce:     nonce,
		r:         r,
		w:         w,
		outgoing:  outgoing,
		incoming:  incoming,
		codec:     DefaultCodec,
		heartbeat: DefaultConfig.heartbeat,
		timeout:   DefaultConfig.timeout,
		hb:        time.NewTimer(DefaultConfig.heartbeat),
	}
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	reg.peers[address] = p
}

// count will return the number of peers currently in the registry.
func (reg *Registry) count() int {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	return len(reg.peers)
}

// get will return the peer with the given address.
func (reg *Registry) get(addr string) *peer {
	reg.mutex.RLock()
	defer reg.mutex.RUnlock()
	return reg.peers[addr]
}

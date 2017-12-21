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
	"errors"
	"net"
	"sync"
)

// Registry presents a registry to manage peers.
type Registry interface {
	Add(peer *Peer) error
	Has(address string) bool
	Get(address string) (*Peer, bool)
	Remove(address string) error
	Count() uint
	List() []string
}

// Peer contains the data on a given peer.
type Peer struct {
	conn   net.Conn
	input  chan interface{}
	output chan interface{}
	nonce  []byte
}

// SimpleRegistry takes care of managing peer data.
type SimpleRegistry struct {
	mutex sync.Mutex
	peers map[string]*Peer
}

// NewSimpleRegistry creates a new peer registry.
func NewSimpleRegistry() *SimpleRegistry {
	return &SimpleRegistry{peers: make(map[string]*Peer)}
}

// Add will add a new peer to the registry.
func (r *SimpleRegistry) Add(peer *Peer) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	address := peer.conn.RemoteAddr().String()
	_, ok := r.peers[address]
	if ok {
		return errors.New("peer already exists")
	}
	r.peers[address] = peer
	return nil
}

// Has will check if we have this peer.
func (r *SimpleRegistry) Has(address string) bool {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, ok := r.peers[address]
	return ok
}

// Get will return the peer with the given address.
func (r *SimpleRegistry) Get(address string) (*Peer, bool) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	peer, ok := r.peers[address]
	return peer, ok
}

// Remove will remove a peer from the registry.
func (r *SimpleRegistry) Remove(address string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	_, ok := r.peers[address]
	if !ok {
		return errors.New("peer not found")
	}
	delete(r.peers, address)
	return nil
}

// Count will return the count of peers in the map.
func (r *SimpleRegistry) Count() uint {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return uint(len(r.peers))
}

// List will give us a list of all peer addresses.
func (r *SimpleRegistry) List() []string {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	addresses := make([]string, 0, len(r.peers))
	for address := range r.peers {
		addresses = append(addresses, address)
	}
	return addresses
}

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
	"errors"
	"sync"

	"github.com/alvalor/alvalor-go/types"
)

type filterFunc func(*peer) bool

func peerHasEntity(has bool, hash types.Hash) func(*peer) bool {
	return func(p *peer) bool {
		_, ok := p.tags[hash]
		return ok == has
	}
}

func peerIsActive(active bool) func(*peer) bool {
	return func(p *peer) bool {
		return p.active == active
	}
}

type peer struct {
	active  bool
	tags    map[types.Hash]struct{}
	pending map[types.Hash]struct{}
}

type simplePeers struct {
	sync.Mutex
	peers map[string]*peer
}

func newPeers() *simplePeers {
	return &simplePeers{
		peers: make(map[string]*peer),
	}
}

func (s *simplePeers) Active(address string) {
	s.Lock()
	defer s.Unlock()

	p, ok := s.peers[address]
	if !ok {
		p = &peer{
			tags:    make(map[types.Hash]struct{}),
			pending: make(map[types.Hash]struct{}),
		}
		s.peers[address] = p
	}
	p.active = true
}

func (s *simplePeers) Inactive(address string) {
	s.Lock()
	defer s.Unlock()

	p, ok := s.peers[address]
	if !ok {
		return
	}
	p.active = false
}

func (s *simplePeers) Requested(address string, hash types.Hash) {
	s.Lock()
	defer s.Unlock()

	p, ok := s.peers[address]
	if !ok {
		return
	}

	p.pending[hash] = struct{}{}
}

func (s *simplePeers) Received(address string, hash types.Hash) {
	s.Lock()
	defer s.Unlock()

	p, ok := s.peers[address]
	if !ok {
		return
	}

	delete(p.pending, hash)
	p.tags[hash] = struct{}{}
}

func (s *simplePeers) NumPending(address string) (uint, error) {
	s.Lock()
	defer s.Unlock()

	p, ok := s.peers[address]
	if !ok {
		return 0, errors.New("peer not found")
	}

	return uint(len(p.pending)), nil
}

// Find will find the peers according to the given filters.
func (s *simplePeers) Find(filters ...filterFunc) []string {
	var addresses []string
Outer:
	for address, p := range s.peers {
		for _, filter := range filters {
			if !filter(p) {
				continue Outer
			}
		}
		addresses = append(addresses, address)
	}
	return addresses
}

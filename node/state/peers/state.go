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

package peers

import (
	"errors"
	"sync"

	"github.com/alvalor/alvalor-go/types"
)

// State represents the state of all peers.
type State struct {
	sync.Mutex
	peers map[string]*Peer
}

// NewState creates a new state for peers.
func NewState() *State {
	return &State{
		peers: make(map[string]*Peer),
	}
}

// Active marks a peer as active.
func (s *State) Active(address string) {
	s.Lock()
	defer s.Unlock()

	p, ok := s.peers[address]
	if !ok {
		p = &Peer{
			yes: make(map[types.Hash]struct{}),
			no:  make(map[types.Hash]struct{}),
		}
		s.peers[address] = p
	}
	p.active = true
}

// Inactive marks a peer as inactive.
func (s *State) Inactive(address string) {
	s.Lock()
	defer s.Unlock()

	p, ok := s.peers[address]
	if !ok {
		return
	}
	p.active = false
}

// Received marks a download as received for a given peer.
func (s *State) Received(address string, hash types.Hash) {
	s.Lock()
	defer s.Unlock()

	p, ok := s.peers[address]
	if !ok {
		return
	}

	p.yes[hash] = struct{}{}
}

// Seen returns a list of entities a peer is aware of.
func (s *State) Seen(address string) ([]types.Hash, error) {
	s.Lock()
	defer s.Unlock()

	p, ok := s.peers[address]
	if !ok {
		return nil, errors.New("peer not found")
	}

	seen := make([]types.Hash, 0, len(p.yes))
	for hash := range p.yes {
		seen = append(seen, hash)
	}

	return seen, nil
}

// Addresses will find the peers according to the given filters.
func (s *State) Addresses(filters ...FilterFunc) []string {
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

// Count will return the count of peers corresponding to the filters.
func (s *State) Count(filters ...FilterFunc) uint {
	return uint(len(s.Addresses(filters...)))
}

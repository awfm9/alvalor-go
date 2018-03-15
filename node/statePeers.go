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
	"sync"

	"github.com/alvalor/alvalor-go/types"
)

type peerManager interface {
	Active(address string)
	Inactive(address string)
	Actives() []string
	Tag(address string, hash types.Hash)
	Tags(hash types.Hash) []string
}

type simplePeers struct {
	sync.Mutex
	actives map[string]bool
	tags    map[types.Hash][]string
}

func newPeers() *simplePeers {
	return &simplePeers{
		actives: make(map[string]bool),
		tags:    make(map[types.Hash][]string),
	}
}

func (s *simplePeers) Active(address string) {
	s.Lock()
	defer s.Unlock()

	s.actives[address] = true
}

func (s *simplePeers) Inactive(address string) {
	s.Lock()
	defer s.Unlock()

	delete(s.actives, address)
}

func (s *simplePeers) Actives() []string {
	s.Lock()
	defer s.Unlock()

	actives := make([]string, 0, len(s.actives))
	for address := range s.actives {
		actives = append(actives, address)
	}

	return actives
}

func (s *simplePeers) Tag(address string, hash types.Hash) {
	s.Lock()
	defer s.Unlock()

	s.tags[hash] = append(s.tags[hash], address)
}

func (s *simplePeers) Tags(hash types.Hash) []string {
	s.Lock()
	defer s.Unlock()

	return s.tags[hash]
}

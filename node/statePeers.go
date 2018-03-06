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

import "sync"

type peerManager interface {
	Active(address string)
	Inactive(address string)
	Actives() []string
	Tag(address string, id []byte)
	Tags(id []byte) []string
}

type simplePeers struct {
	sync.Mutex
	actives map[string]bool
	tags    map[string][]string
}

func newPeers() *simplePeers {
	return &simplePeers{
		actives: make(map[string]bool),
		tags:    make(map[string][]string),
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

func (s *simplePeers) Tag(address string, id []byte) {
	s.Lock()
	defer s.Unlock()

	s.tags[string(id)] = append(s.tags[string(id)], address)
}

func (s *simplePeers) Tags(id []byte) []string {
	s.Lock()
	defer s.Unlock()

	return s.tags[string(id)]
}

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

type stateManager interface {
	On(address string)
	Off(address string)
	Active() []string
	Tag(address string, id []byte)
	Seen(id []byte) []string
}

type simpleStateManager struct {
	active map[string]bool
	seen   map[string][]string
}

func newSimpleStateManager() *simpleStateManager {
	return &simpleStateManager{
		active: make(map[string]bool),
		seen:   make(map[string][]string),
	}
}

func (s *simpleStateManager) On(address string) {
	s.active[address] = true
}

func (s *simpleStateManager) Off(address string) {
	delete(s.active, address)
}

func (s *simpleStateManager) Active() []string {
	active := make([]string, 0, len(s.active))
	for address := range s.active {
		active = append(active, address)
	}
	return active
}

func (s *simpleStateManager) Tag(address string, id []byte) {
	s.seen[string(id)] = append(s.seen[string(id)], address)
}

func (s *simpleStateManager) Seen(id []byte) []string {
	return s.seen[string(id)]
}

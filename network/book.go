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
	"math"
	"math/rand"
	"sort"
	"sync"
)

// Book interface.
type Book interface {
	Add(addr string)
	Whitelist(addr string)
	Blacklist(addr string)
	Connected(addr string)
	Disconnected(addr string)
	Dropped(addr string)
	Failed(addr string)
	Get() (string, error)
	Sample() ([]string, error)
}

// DefaultBook variable.
var DefaultBook = &SimpleBook{
	blacklist:  make(map[string]struct{}),
	entries:    make(map[string]*entry),
	sampleSize: 10,
}

// entry struct.
type entry struct {
	addr    string
	active  bool
	success int
	failure int
}

// score method.
func (e entry) score() float64 {
	if e.failure == 0 {
		return 1
	}
	score := float64(e.success) / float64(e.failure)
	return math.Max(score/100, 1)
}

// error variables.
var (
	errAddrInvalid = errors.New("invalid address")
	errAddrUnknown = errors.New("unknown address")
	errBookEmpty   = errors.New("book empty")
)

// SimpleBook struct.
type SimpleBook struct {
	mutex      sync.Mutex
	blacklist  map[string]struct{}
	entries    map[string]*entry
	sampleSize int
}

// NewSimpleBook function.
func NewSimpleBook() *SimpleBook {
	return &SimpleBook{
		blacklist:  make(map[string]struct{}),
		entries:    make(map[string]*entry),
		sampleSize: 10,
	}
}

// Whitelist method.
func (s *SimpleBook) Whitelist(addr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.blacklist, addr)
	peer, ok := s.entries[addr]
	if !ok {
		return
	}
	peer.failure = 0
	peer.success = 1
}

// Blacklist method.
func (s *SimpleBook) Blacklist(addr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.entries, addr)
	s.blacklist[addr] = struct{}{}
}

// Add method.
func (s *SimpleBook) Add(addr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.blacklist[addr]
	if ok {
		return
	}
	s.entries[addr] = &entry{addr: addr}
}

// Connected method.
func (s *SimpleBook) Connected(addr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[addr]
	if !ok {
		return
	}
	e.active = true
	e.success++
}

// Disconnected method.
func (s *SimpleBook) Disconnected(addr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[addr]
	if !ok {
		return
	}
	e.active = false
}

// Dropped method.
func (s *SimpleBook) Dropped(addr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[addr]
	if !ok {
		return
	}
	e.active = false
	e.failure++
}

// Failed method.
func (s *SimpleBook) Failed(addr string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[addr]
	if !ok {
		return
	}
	e.active = false
	e.failure++
}

// Get method.
func (s *SimpleBook) Get() (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	entries := s.slice(false)
	if len(entries) == 0 {
		return "", errBookEmpty
	}
	sort.Sort(byPriority(entries))
	e := entries[0]
	e.active = true
	return e.addr, nil
}

// Sample method.
func (s *SimpleBook) Sample() ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	entries := s.slice(true)
	if len(entries) == 0 {
		return nil, errors.New("no valid addresses")
	}
	if len(entries) > s.sampleSize {
		for i := 0; i < len(entries); i++ {
			j := rand.Intn(i + 1)
			entries[i], entries[j] = entries[j], entries[i]
		}
		entries = entries[:s.sampleSize]
	}
	addrs := make([]string, 0, s.sampleSize)
	for _, e := range entries {
		addrs = append(addrs, e.addr)
	}
	return addrs, nil
}

// slice method.
func (s *SimpleBook) slice(active bool) []*entry {
	entries := make([]*entry, 0)
	for _, e := range s.entries {
		if !active && e.active {
			continue
		}
		entries = append(entries, e)
	}
	return entries
}

// byPriority type.
type byPriority []*entry

// Len method.
func (b byPriority) Len() int {
	return len(b)
}

// Less method.
func (b byPriority) Less(i int, j int) bool {
	return b[i].score() < b[j].score()
}

// Swap method.
func (b byPriority) Swap(i int, j int) {
	b[i], b[j] = b[j], b[i]
}

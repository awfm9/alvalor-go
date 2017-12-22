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

package book

import (
	"errors"
	"sort"
	"sync"
)

// Simple is an address book with simple scoring.
type Simple struct {
	mutex     sync.Mutex
	blacklist map[string]struct{}
	entries   map[string]*Entry
}

// enumeration of different errors that we can return from address book functions.
var (
	errAddrInvalid  = errors.New("invalid address")
	errAddrUnknown  = errors.New("unknown address")
	errBookEmpty    = errors.New("book empty")
	errInvalidCount = errors.New("invalid address count")
)

// NewSimple creates a new default initialized instance of a simple address book.
func NewSimple() *Simple {
	return &Simple{
		blacklist: make(map[string]struct{}),
		entries:   make(map[string]*Entry),
	}
}

// Add will add an address to the list of available peer addresses, unless it is blacklisted.
func (s *Simple) Add(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	_, ok := s.blacklist[address]
	if ok {
		return
	}
	s.entries[address] = &Entry{Address: address}
}

// Invalid should be called whenever an address should be considered permanently to be an
// invalid peer.
func (s *Simple) Invalid(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.entries, address)
	s.blacklist[address] = struct{}{}
}

// Error should be called whenever there is an error on a connection that could be
// temporary in nature.
func (s *Simple) Error(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[address]
	if !ok {
		return
	}
	e.Failure++
}

// Success should be called by consumers whenever a successful connection to the peer with the
// given address was established. It is used to keep track of the active status and to increase the
// success count of the peer.
func (s *Simple) Success(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[address]
	if !ok {
		return
	}
	e.Active = true
	e.Success++
}

// Dropped should be called by consumers whenever a peer was disconnected. It is
// used to keep track of the active status.
func (s *Simple) Dropped(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[address]
	if !ok {
		return
	}
	e.Active = false
}

// Failure should be called whenever connection to a peer failed. It is used to keep track of the
// failure & active status of a peer.
func (s *Simple) Failure(address string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	e, ok := s.entries[address]
	if !ok {
		return
	}
	e.Active = false
	e.Failure++
}

// Sample will return entries limited by count, filtered by specified filter function and sorted by specified sort function
func (s *Simple) Sample(count uint, params ...interface{}) ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// extract all the parameters for the sample
	var filters []filterFunc
	var sorts []sortFunc
	for _, param := range params {
		switch f := param.(type) {
		case filterFunc:
			filters = append(filters, f)
		case sortFunc:
			sorts = append(sorts, f)
		}
	}

	// check if we have a valid count
	if count == 0 {
		return nil, errInvalidCount
	}

	// apply the filter
	var entries []*Entry
	for _, e := range s.entries {
		for _, filter := range filters {
			if !filter(e) {
				continue
			}
		}
		entries = append(entries, e)
	}

	// check if we have any entries that fulfill the criteria
	if len(entries) == 0 {
		return nil, errBookEmpty
	}

	// sort the entries
	sort.Slice(entries, func(i int, j int) bool {
		for _, less := range sorts {
			if less(entries[i], entries[j]) {
				return true
			}
		}
		return false
	})

	// make sure we don't return more than count
	if uint(len(entries)) > count {
		entries = entries[:count]
	}

	// build slice of just addresses
	addresses := make([]string, 0, count)
	for _, e := range entries {
		addresses = append(addresses, e.Address)
	}

	return addresses, nil
}

// slice method.
func (s *Simple) slice(filter func(*Entry) bool) []*Entry {
	entries := make([]*Entry, 0)
	for _, e := range s.entries {
		if !filter(e) {
			continue
		}
		entries = append(entries, e)
	}
	return entries
}

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
// GNU Affero General Public License for more detailb.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package network

import (
	"sort"
	"sync"
)

// Book is an address book with simple scoring.
type Book struct {
	mutex     sync.Mutex
	blacklist map[string]struct{}
	entries   map[string]*entry
}

// NewBook creates a new default initialized instance of a simple address book.
func NewBook() *Book {
	return &Book{
		blacklist: make(map[string]struct{}),
		entries:   make(map[string]*entry),
	}
}

// Found will add an address to the list of available peer addresses, unless it is blacklisted.
func (b *Book) Found(address string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	_, ok := b.blacklist[address]
	if ok {
		return
	}
	b.entries[address] = &entry{Address: address}
}

// Invalid should be called whenever an address should be considered permanently to be an
// invalid peer.
func (b *Book) Invalid(address string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.entries, address)
	b.blacklist[address] = struct{}{}
}

// Error should be called whenever there is an error on a connection that could be
// temporary in nature.
func (b *Book) Error(address string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	e, ok := b.entries[address]
	if !ok {
		return
	}
	e.Failure++
}

// Success should be called by consumers whenever a successful connection to the peer with the
// given address was established. It is used to keep track of the active status and to increase the
// success count of the peer.
func (b *Book) Success(address string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	e, ok := b.entries[address]
	if !ok {
		return
	}
	e.Active = true
	e.Success++
}

// Dropped should be called by consumers whenever a peer was disconnected. It is
// used to keep track of the active status.
func (b *Book) Dropped(address string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	e, ok := b.entries[address]
	if !ok {
		return
	}
	e.Active = false
}

// Failure should be called whenever connection to a peer failed. It is used to keep track of the
// failure & active status of a peer.
func (b *Book) Failure(address string) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	e, ok := b.entries[address]
	if !ok {
		return
	}
	e.Active = false
	e.Failure++
}

// Sample will return entries limited by count, filtered by specified filter function and sorted by specified sort function
func (b *Book) Sample(count uint, params ...interface{}) []string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	// extract all the parameters for the sample
	var filters []func(e *entry) bool
	var sorts []func(*entry, *entry) bool
	for _, param := range params {
		switch f := param.(type) {
		case func(e *entry) bool:
			filters = append(filters, f)
		case func(*entry, *entry) bool:
			sorts = append(sorts, f)
		}
	}

	// apply the filter
	var entries []*entry
Outer:
	for _, e := range b.entries {
		for _, filter := range filters {
			if !filter(e) {
				continue Outer
			}
		}
		entries = append(entries, e)
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

	return addresses
}

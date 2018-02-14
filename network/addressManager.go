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
	"sort"
	"sync"
)

type addressManager interface {
	Add(address string)
	Remove(address string)
	Block(address string)
	Unblock(address string)
	Sample(count uint, params ...interface{}) []string
}

type simpleAddressManager struct {
	sync.Mutex
	blacklist map[string]bool
	addresses map[string]bool
}

func newSimpleAddressManager() *simpleAddressManager {
	return &simpleAddressManager{
		blacklist: make(map[string]bool),
		addresses: make(map[string]bool),
	}
}

func (am *simpleAddressManager) Add(address string) {
	am.Lock()
	defer am.Unlock()
	am.addresses[address] = true
}

func (am *simpleAddressManager) Remove(address string) {
	am.Lock()
	defer am.Unlock()
	delete(am.addresses, address)
}

func (am *simpleAddressManager) Block(address string) {
	am.Lock()
	defer am.Unlock()
	am.blacklist[address] = true
}

func (am *simpleAddressManager) Unblock(address string) {
	am.Lock()
	defer am.Unlock()
	delete(am.blacklist, address)
}

func (am *simpleAddressManager) Sample(count uint, params ...interface{}) []string {
	am.Lock()
	defer am.Unlock()

	// initialize list of filters & sorts
	var filters []func(string) bool
	var sorts []func(string, string) bool

	// add blacklist filter
	blacklist := func(address string) bool {
		return !am.blacklist[address]
	}
	filters = append(filters, blacklist)

	// add custom filters & sorts
	for _, param := range params {
		switch f := param.(type) {
		case func(string) bool:
			filters = append(filters, f)
		case func(string, string) bool:
			sorts = append(sorts, f)
		}
	}

	// apply the filters
	var addresses []string
Outer:
	for address := range am.addresses {
		for _, filter := range filters {
			if !filter(address) {
				continue Outer
			}
		}
		addresses = append(addresses, address)
	}

	// prioritize whitelisted entries
	sort.Slice(addresses, func(i int, j int) bool {
		for _, less := range sorts {
			if less(addresses[i], addresses[j]) {
				return true
			}
			if less(addresses[j], addresses[i]) {
				return false
			}
		}
		return false
	})

	// limit count
	if uint(len(addresses)) > count {
		addresses = addresses[:count]
	}

	return addresses
}

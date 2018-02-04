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
	"sync"
)

type pendingManager interface {
	Claim(address string) error
	Release(address string) error
	Addresses() []string
	Count() uint
}

type simplePendingManager struct {
	sync.Mutex
	max       uint
	addresses map[string]struct{}
}

func newSimplePendingManager(max uint) *simplePendingManager {
	return &simplePendingManager{
		max:       max,
		addresses: make(map[string]struct{}, max),
	}
}

// Claim will reduce the amount of available pending by one.
func (pending *simplePendingManager) Claim(address string) error {
	pending.Lock()
	defer pending.Unlock()
	if uint(len(pending.addresses)) >= pending.max {
		return errors.New("no free pending")
	}
	_, ok := pending.addresses[address]
	if ok {
		return errors.New("address already pending")
	}
	pending.addresses[address] = struct{}{}
	return nil
}

// Release will increase the amount of available pending by one.
func (pending *simplePendingManager) Release(address string) error {
	pending.Lock()
	defer pending.Unlock()
	_, ok := pending.addresses[address]
	if !ok {
		return errors.New("address not pending")
	}
	delete(pending.addresses, address)
	return nil
}

// Addresses will return the addresses currently pending connection.
func (pending *simplePendingManager) Addresses() []string {
	pending.Lock()
	defer pending.Unlock()
	addresses := make([]string, 0, len(pending.addresses))
	for address := range pending.addresses {
		addresses = append(addresses, address)
	}
	return addresses
}

// Count will return the number of currently pending pending.
func (pending *simplePendingManager) Count() uint {
	pending.Lock()
	defer pending.Unlock()
	return uint(len(pending.addresses))
}

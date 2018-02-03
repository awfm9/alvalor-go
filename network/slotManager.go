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

type slotManager interface {
	Claim() error
	Release() error
	Pending() uint
}

type simpleSlotManager struct {
	sync.Mutex
	initial   uint
	available uint
}

func newSimpleSlotManager(slots uint) *simpleSlotManager {
	return &simpleSlotManager{initial: slots, available: slots}
}

// Claim will reduce the amount of available slots by one.
func (slots *simpleSlotManager) Claim() error {
	slots.Lock()
	defer slots.Unlock()
	if slots.available == 0 {
		return errors.New("no free slots")
	}
	slots.available--
	return nil
}

// Release will increase the amount of available slots by one.
func (slots *simpleSlotManager) Release() error {
	slots.Lock()
	defer slots.Unlock()
	if slots.available == slots.initial {
		return errors.New("all slots released")
	}
	slots.available++
	return nil
}

// Pending will return the number of currently pending slots.
func (slots *simpleSlotManager) Pending() uint {
	slots.Lock()
	defer slots.Unlock()
	return slots.initial - slots.available
}

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

package event

import (
	"errors"
	"time"

	"github.com/alvalor/alvalor-go/types"
)

// Manager represents a manager for event notifications.
type Manager struct {
	stream chan<- interface{}
}

// NewManager creates a new event manager.
func NewManager(stream chan<- interface{}) *Manager {
	return &Manager{stream: stream}
}

// Header triggers a new header event.
func (mgr *Manager) Header(hash types.Hash) error {
	select {
	case mgr.stream <- Header{hash: hash}:
	case <-time.After(10 * time.Millisecond):
		return errors.New("subscriber stalling")
	}
	return nil
}

// Transaction creates a new transaction event.
func (mgr *Manager) Transaction(hash types.Hash) error {
	select {
	case mgr.stream <- Transaction{hash: hash}:
	case <-time.After(10 * time.Millisecond):
		return errors.New("subscriber stalling")
	}
	return nil
}

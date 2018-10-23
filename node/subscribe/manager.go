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

package subscribe

import (
	"errors"
	"time"

	"github.com/alvalor/alvalor-go/types"
)

// Manager represents a manager for event notifications.
type Manager struct {
	stream  chan interface{}
	timeout time.Duration
	subs    map[chan<- interface{}][]func(interface{}) bool
}

// NewManager creates a new event manager.
func NewManager(buffer uint, timeout time.Duration) *Manager {
	mgr := &Manager{
		stream:  make(chan interface{}, buffer),
		timeout: timeout,
		subs:    make(map[chan<- interface{}][]func(interface{}) bool),
	}
	return mgr
}

// Subscribe adds a subscriber to the event output.
func (mgr *Manager) Subscribe(sub chan<- interface{}, filters ...func(interface{}) bool) {
	mgr.subs[sub] = filters
}

// Unsubscribe removes a subscriber from the event output.
func (mgr *Manager) Unsubscribe(sub chan<- interface{}) {
	delete(mgr.subs, sub)
}

// Header triggers a new header event.
func (mgr *Manager) Header(hash types.Hash) error {
	return mgr.event(Header{hash: hash})
}

// Transaction creates a new transaction event.
func (mgr *Manager) Transaction(hash types.Hash) error {
	return mgr.event(Transaction{hash: hash})
}

// event submits the event to the channel.
func (mgr *Manager) event(event interface{}) error {
	select {
	case mgr.stream <- event:
	case <-time.After(mgr.timeout):
		return errors.New("subscriber stalling")
	}
	return nil
}

// TODO: add processing for each subscriber (see stream.go)

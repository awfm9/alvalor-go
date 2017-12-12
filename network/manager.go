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
	"sync"

	"github.com/rs/zerolog"
)

// Manager represents a manager for events, executing the respective actions we
// want depending on the events.
type Manager struct {
	log    zerolog.Logger
	events <-chan interface{}
}

// NewManager creathes a new manager of network events.
func NewManager(log zerolog.Logger, events <-chan interface{}) *Manager {
	mgr := &Manager{
		log:    log,
		events: events,
	}
	return mgr
}

// process will launch the processing of the processor.
func (mgr *Manager) process(wg *sync.WaitGroup) {

	wg.Add(1)

	for event := range mgr.events {
		mgr.log.Debug().Interface("event", event).Msg("processing event")
		switch event.(type) {
		case Tick:
			// process tick
			//
			// tick -> rebalance peers
		case Command:
			// process command
			//
			// send -> send message to desired peer
		case Network:
			// process network
			//
			// connect -> perform handshake
			// handshaked -> log in book, add to registry, bubble up
			// disconnect -> log in book, remove from registry, bubble up event
			// failed -> log in book
			// error -> log in book
			// message -> handle message
			//   ping -> pong
			//   discover -> peers
			//   pong -> noop
			//   peers -> (add to book)
		}
	}

	wg.Done()
}

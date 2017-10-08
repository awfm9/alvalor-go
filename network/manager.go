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
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// Manager represents a manager for events, executing the respective actions we
// want depending on the events.
type Manager struct {
	log        *zap.Logger
	wg         *sync.WaitGroup
	book       Book
	node       Node
	reg        Registry
	events     <-chan interface{}
	addresses  chan<- string
	subscriber chan<- interface{}
	running    uint32
	minPeers   uint
	maxPeers   uint
}

// NewManager creates a new manager of network events.
func NewManager(log *zap.Logger, wg *sync.WaitGroup, book Book, events <-chan interface{}, addresses chan<- string, subscriber chan<- interface{}, options ...func(*Manager)) *Manager {
	mgr := &Manager{
		log:        log,
		wg:         wg,
		book:       book,
		events:     events,
		addresses:  addresses,
		subscriber: subscriber,
		running:    1,
		minPeers:   4,
		maxPeers:   8,
	}
	for _, option := range options {
		option(mgr)
	}
	wg.Add(1)
	go mgr.process()
	return mgr
}

// process will launch the processing of the processor.
func (mgr *Manager) process() {

Loop:
	for atomic.LoadUint32(&mgr.running) > 0 {

		// make sure we re-enter the loop every second to check for shutdown
		var event interface{}
		select {
		case event = <-mgr.events:
		case <-time.After(100 * time.Millisecond):
			continue Loop
		}

		// depending on the event, execute related actions
		switch e := event.(type) {
		case Balance:
			add := e.Max - mgr.reg.count()
			addresses, err := mgr.book.Sample(add, IsActive(false), ByPrioritySort())
			if err != nil {
				mgr.log.Info("not enough addresses in book", zap.Error(err))
			}
			for _, address := range addresses {
				mgr.addresses <- address
			}
		case Disconnection:
			mgr.book.Disconnected(e.Address)
			mgr.reg.remove(e.Address)
			mgr.subscriber <- e
		case Failure:
			mgr.book.Failed(e.Address)
		case Violation:
			mgr.book.Blacklist(e.Address)
		case Message:
			mgr.subscriber <- e
		case Connection:
			mgr.book.Connected(e.Address)
			mgr.reg.add(e.Address, e.Conn, e.Nonce)
			mgr.subscriber <- e
		default:
			mgr.log.Error("invalid network event", zap.String("type", fmt.Sprintf("%T", e)))
		}
	}

	// before finishing shutdown, close the channels we are producing for
	close(mgr.addresses)
	close(mgr.subscriber)

	// let the waitgroup know we are done
	mgr.wg.Done()
}

// Close will shut down the manager.
func (mgr *Manager) Close() {
	mgr.running = 0
}

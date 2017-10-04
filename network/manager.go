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
	minPeers   uint
	maxPeers   uint
}

// NewManager creates a new manager of network events.
func NewManager(log *zap.Logger, wg *sync.WaitGroup, book Book, events <-chan interface{}, addresses chan<- string, subscriber chan<- interface{}, options ...func(*Manager)) *Manager {
	mgr := &Manager{
		log:    log,
		wg:     wg,
		book:   book,
		events: events,
	}
	for _, option := range options {
		option(mgr)
	}
	go mgr.process()
	return mgr
}

// process will launch the processing of the processor.
func (mgr *Manager) process() {
	for event := range mgr.events {
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
	close(mgr.addresses)
	close(mgr.subscriber)
	mgr.wg.Done()
}

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
	"time"

	"github.com/pierrec/lz4"
	"go.uber.org/zap"
)

// Manager represents a manager for events, executing the respective actions we
// want depending on the events.
type Manager struct {
	log        *zap.Logger
	book       Book
	node       Node
	events     <-chan interface{}
	subscriber chan<- interface{}
}

// NewManager creates a new manager of network events.
func NewManager(log *zap.Logger, book Book, events <-chan interface{}, subscriber chan<- interface{}) *Manager {
	mgr := &Manager{
		log:    log,
		book:   book,
		events: events,
	}
	return mgr
}

// Process will launch the processing of the processor.
func (mgr *Manager) Process() {
	for event := range mgr.events {
		switch e := event.(type) {
		case Disconnection:
			mgr.book.Disconnected(e.Address)
			mgr.subscriber <- e
		case Failure:
			mgr.book.Failed(e.Address)
		case Violation:
			mgr.book.Blacklist(e.Address)
		case Message:
			mgr.subscriber <- e
		case Connection:
			mgr.book.Connected(e.Address)
			r := lz4.NewReader(e.Conn)
			w := lz4.NewWriter(e.Conn)
			outgoing := make(chan interface{}, 16)
			incoming := make(chan interface{}, 16)
			p := &peer{
				conn:      e.Conn,
				addr:      e.Address,
				nonce:     e.Nonce,
				r:         r,
				w:         w,
				outgoing:  outgoing,
				incoming:  incoming,
				codec:     DefaultCodec,
				heartbeat: DefaultConfig.heartbeat,
				timeout:   DefaultConfig.timeout,
				hb:        time.NewTimer(DefaultConfig.heartbeat),
			}
			_ = p
			mgr.subscriber <- e
		default:
			mgr.log.Error("invalid network event", zap.String("type", fmt.Sprintf("%T", e)))
		}
	}
}

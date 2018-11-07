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
	"sync"

	"github.com/alvalor/alvalor-go/network"
	"github.com/alvalor/alvalor-go/node/message"
	"github.com/rs/zerolog"
)

// Handler represents the handler for events received from the network layer.
type Handler struct {
	log     zerolog.Logger
	net     Network
	headers Headers
	peers   Peers
	message Message
}

// Process makes the event handler process an event.
func (handler *Handler) Process(wg *sync.WaitGroup, event interface{}) {
	wg.Add(1)
	go handler.process(wg, event)
}

func (handler *Handler) process(wg *sync.WaitGroup, event interface{}) {
	defer wg.Done()

	// configure logger
	log := handler.log.With().Str("component", "event").Logger()
	log.Debug().Msg("event routine started")
	defer log.Debug().Msg("event routine stopped")

	switch e := event.(type) {

	case network.Connected:

		handler.peers.Active(e.Address)

		// send our current best distance
		_, distance := handler.headers.Path()
		status := &message.Status{
			Distance: distance,
		}
		err := handler.net.Send(e.Address, status)
		if err != nil {
			log.Error().Err(err).Msg("could not send status message")
			return
		}

	case network.Disconnected:
		handler.peers.Inactive(e.Address)

	case network.Received:
		handler.message.Process(e.Address, e.Message)
	}
}

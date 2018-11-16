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
	switch e := event.(type) {
	case network.Connected:
		go handler.processConnected(wg, e)
	case network.Disconnected:
		go handler.processDisconnected(wg, e)
	case network.Received:
		go handler.processReceived(wg, e)
	}
}

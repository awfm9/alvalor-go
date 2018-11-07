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
)

func (handler *Handler) processReceived(wg *sync.WaitGroup, received network.Received) {
	defer wg.Done()

	// configure logger
	with := handler.log.With()
	with.Str("component", "event")
	with.Str("event_type", "received")
	with.Str("address", received.Address)
	log := with.Logger()

	// wrap routine in start and stop messages
	log.Debug().Msg("routine started")
	defer log.Debug().Msg("routine stopped")

	handler.message.Process(wg, received.Address, received.Message)
}

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

func (handler *Handler) processDisconnected(wg *sync.WaitGroup, disconnected network.Disconnected) {
	defer wg.Done()

	// configure logger
	log := handler.log.With().Str("component", "event_disconnected").Logger()
	log.Debug().Msg("event_disconnected routine started")
	defer log.Debug().Msg("event_disconnected routine stopped")

	handler.peers.Inactive(disconnected.Address)
}

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

package node

import (
	"sync"

	"github.com/rs/zerolog"

	"github.com/alvalor/alvalor-go/network"
)

func handleReceiving(log zerolog.Logger, wg *sync.WaitGroup, subscription <-chan interface{}, handlers handlerManager, state stateManager) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "receiver").Logger()
	log.Info().Msg("receiving routine started")
	defer log.Info().Msg("receiving routine stopped")

	for event := range subscription {
		switch e := event.(type) {

		// in the case of a connected event, we start tracking the peer state
		case *network.Connected:
			log.Info().Str("address", e.Address).Msg("peer connected")
			state.Active(e.Address)

		// in the case of a disconnected event, we stop tracking the peer state
		case *network.Disconnected:
			log.Info().Str("address", e.Address).Msg("peer disconnected")
			state.Inactive(e.Address)

		// in the case of a received event, we start processing the peer message
		case *network.Received:
			log.Info().Str("address", e.Address).Msg("message received")
			entity, ok := e.Message.(Entity)
			if !ok {
				log.Error().Msg("received event is not entity")
				continue
			}
			state.Tag(e.Address, entity.ID())
			handlers.Process(entity)
		}
	}
}

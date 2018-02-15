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
	"encoding/hex"
	"sync"

	"github.com/rs/zerolog"
)

func handlePropagating(log zerolog.Logger, wg *sync.WaitGroup, entity Hasher, state stateManager, net networkManager) {
	defer wg.Done()

	var (
		hash = entity.Hash()
	)

	// configure logger
	log = log.With().Str("component", "propagator").Str("hash", hex.EncodeToString(hash)).Logger()
	log.Info().Msg("propagating routine started")
	defer log.Info().Msg("propagating routine stopped")

	// create lookup to know who to exclude from broadcast
	known := state.Set(hash)
	lookup := make(map[string]struct{}, len(known))
	for _, address := range known {
		lookup[address] = struct{}{}
	}

	// send it to each peer not excluded
	peers := net.Peers()
	for _, address := range peers {
		_, ok := lookup[address]
		if ok {
			continue
		}
		err := net.Send(address, entity)
		if err != nil {
			log.Error().Err(err).Str("address", address).Msg("could not propagate entity")
			continue
		}
	}
}

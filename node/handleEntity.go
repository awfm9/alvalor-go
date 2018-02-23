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

	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
)

func handleEntity(log zerolog.Logger, wg *sync.WaitGroup, net Network, state stateManager, entity Entity) {
	defer wg.Done()

	var (
		id = entity.ID()
	)

	// configure logger
	log = log.With().Str("component", "entity").Str("id", hex.EncodeToString(id)).Logger()
	log.Debug().Msg("entity routine started")
	defer log.Debug().Msg("entity routine stopped")

	switch entity.(type) {

	case *types.Transaction:

		// create lookup to know who to exclude from broadcast
		tags := state.Tags(id)
		lookup := make(map[string]struct{}, len(tags))
		for _, address := range tags {
			lookup[address] = struct{}{}
		}

		// for each active peer
		actives := state.Actives()
		for _, address := range actives {

			// skip if he already knows the transaction
			_, ok := lookup[address]
			if ok {
				continue
			}

			// otherwise, send the transaction
			err := net.Send(address, entity)
			if err != nil {
				log.Error().Err(err).Str("address", address).Msg("could not propagate entity")
				continue
			}
		}
	}
}

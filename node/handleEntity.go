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

	"github.com/alvalor/alvalor-go/types"
)

func handleEntity(log zerolog.Logger, wg *sync.WaitGroup, net Network, finder pathfinder, peers peerManager, pool poolManager, entity Entity, events eventManager) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "entity").Logger()
	log.Debug().Msg("entity routine started")
	defer log.Debug().Msg("entity routine stopped")

	switch e := entity.(type) {

	case *types.Header:

		log = log.With().Str("entity_type", "header").Hex("hash", e.Hash[:]).Logger()

		ok := finder.Knows(e.Hash)
		if ok {
			log.Debug().Msg("header already known")
			return
		}

		err := finder.Add(e)
		if err != nil {
			log.Error().Err(err).Msg("could not add header")
			return
		}

		events.Header(e.Hash)

		log.Debug().Msg("header processed")

	case *types.Transaction:

		log = log.With().Str("entity_type", "transaction").Hex("hash", e.Hash[:]).Logger()

		// check if we already know the transaction; if so, ignore it
		ok := pool.Known(e.Hash)
		if ok {
			log.Debug().Msg("transaction already known")
			return
		}

		// add the transaction to the transaction pool
		err := pool.Add(e)
		if err != nil {
			log.Error().Err(err).Msg("could not add transaction")
			return
		}

		events.Transaction(e.Hash)

		// create lookup to know who to exclude from broadcast
		tags := peers.Tags(e.Hash)
		lookup := make(map[string]struct{}, len(tags))
		for _, address := range tags {
			lookup[address] = struct{}{}
		}

		// for each active peer
		actives := peers.Actives()
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

		log.Debug().Msg("transaction processed")
	}
}

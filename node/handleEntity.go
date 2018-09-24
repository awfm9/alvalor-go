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

func handleEntity(log zerolog.Logger, wg *sync.WaitGroup, net Network, finder pathfinder, peers peerManager, pool poolManager, download downloader, entity Entity, events eventManager, handlers Handlers) {
	defer wg.Done()

	// precompute the entity hash
	hash := entity.GetHash()

	// configure logger
	log = log.With().Str("component", "entity").Hex("hash", hash[:]).Logger()
	log.Debug().Msg("entity routine started")
	defer log.Debug().Msg("entity routine stopped")

	switch e := entity.(type) {

	// When we receive a new header, we want to add it to our pathfinder to see
	// whether it creates a better path of total difficulty. If that's the case,
	// we need to synchronize the blocks on that path. This implies canceling
	// transaction downloads for all headers that are no longer on the best path,
	// and starting transaction downloads for all new headers on the best path.
	case *types.Header:

		e.Hash = hash

		log = log.With().Str("entity_type", "header").Logger()

		// if we already know the header, we ignore it
		ok := finder.Knows(e.Hash)
		if ok {
			log.Debug().Msg("header already known")
			return
		}

		// check the validity of the header
		// TODO

		// add the header to the pathfinder
		err := finder.Add(e)
		if err != nil {
			log.Error().Err(err).Msg("could not add header")
			return
		}

		// we let subscribers know that we received a new header
		events.Header(e.Hash)

		// we should propagate it to peers who are unaware of the header
		peers := peers.Tags(e.Hash)
		err = net.Broadcast(e, peers...)
		if err != nil {
			log.Error().Err(err).Msg("could not propagate entity")
			return
		}

		// switch the downloader to the new best path
		path, _ := finder.Longest()
		err = download.Follow(path)
		if err != nil {
			log.Error().Err(err).Msg("could not follow changed path")
			return
		}

		log.Debug().Msg("header processed")

	case *types.Transaction:

		e.Hash = hash

		log = log.With().Str("entity_type", "transaction").Logger()

		// check if we already know the transaction; if so, ignore it
		ok := pool.Knows(e.Hash)
		if ok {
			log.Debug().Msg("transaction already known")
			return
		}

		// check the validity of the transaction
		// TODO

		// add the transaction to the transaction pool
		err := pool.Add(e)
		if err != nil {
			log.Error().Err(err).Msg("could not add transaction")
			return
		}

		events.Transaction(e.Hash)

		// create lookup to know who to exclude from broadcast
		peers := peers.Tags(e.Hash)
		err = net.Broadcast(e, peers...)
		if err != nil {
			log.Error().Err(err).Msg("could not propagate entity")
			return
		}

		log.Debug().Msg("transaction processed")
	}
}

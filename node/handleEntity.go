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

		// otherwise, we try to add it to our header manager
		err := finder.Add(e)
		if err != nil {
			log.Error().Err(err).Msg("could not add header")
			return
		}

		// we let subscribers know that we received a new header
		events.Header(e.Hash)

		// finally, we should propagate it to peers who don't know it
		peers := peers.Tags(e.Hash)
		err = net.Broadcast(e, peers...)
		if err != nil {
			log.Error().Err(err).Msg("could not propagate entity")
			return
		}

		// set the new longest path with the downloader
		path, _ := finder.Longest()
		download.Follow(path)

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

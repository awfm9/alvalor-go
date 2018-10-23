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

package entity

import (
	"sync"

	"github.com/alvalor/alvalor-go/node/peer"
	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
)

// Handler is the handler for entities. We use a struct rather than a
// function so we can mock it easier for testing.
type Handler struct {
	wg           *sync.WaitGroup
	log          zerolog.Logger
	net          Network
	paths        Paths
	events       Events
	headers      Headers
	transactions Transactions
	peers        Peers
}

// NewHandler creates a new handler for entities.
func NewHandler(wg *sync.WaitGroup, log zerolog.Logger, net Network, paths Paths, events Events, headers Headers, transactions Transactions, peers Peers) *Handler {
	handler := &Handler{
		wg:           wg,
		log:          log,
		net:          net,
		paths:        paths,
		events:       events,
		headers:      headers,
		transactions: transactions,
		peers:        peers,
	}
	return handler
}

// Process is the entity handler's function for processing a new entity.
func (handler *Handler) Process(entity types.Entity) {
	handler.wg.Add(1)
	defer handler.wg.Done()

	// precompute the entity hash
	hash := entity.GetHash()

	// configure logger
	log := handler.log.With().Str("component", "entity").Hex("hash", hash[:]).Logger()
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
		ok := handler.headers.Has(e.Hash)
		if ok {
			log.Debug().Msg("header already known")
			return
		}

		// check the validity of the header
		// TODO

		// add the header to the pathfinder
		err := handler.headers.Add(e)
		if err != nil {
			log.Error().Err(err).Msg("could not add header")
			return
		}

		// we let subscribers know that we received a new header
		handler.events.Header(e.Hash)

		// we should propagate it to peers who are unaware of the header
		// TODO: change broadcast to have target addresses and not exclusion
		addresses := handler.peers.Addresses(peer.HasEntity(false, e.Hash))
		err = handler.net.Broadcast(e, addresses...)
		if err != nil {
			log.Error().Err(err).Msg("could not propagate entity")
			return
		}

		// switch the downloader to the new best path
		path, _ := handler.headers.Path()
		err = handler.paths.Follow(path)
		if err != nil {
			log.Error().Err(err).Msg("could not follow changed path")
			return
		}

		log.Debug().Msg("header processed")

	case *types.Transaction:

		e.Hash = hash

		log = log.With().Str("entity_type", "transaction").Logger()

		// check if we already know the transaction; if so, ignore it
		ok := handler.transactions.Has(e.Hash)
		if ok {
			log.Debug().Msg("transaction already known")
			return
		}

		// check the validity of the transaction
		// TODO

		// add the transaction to the transaction pool
		err := handler.transactions.Add(e)
		if err != nil {
			log.Error().Err(err).Msg("could not add transaction")
			return
		}

		handler.events.Transaction(e.Hash)

		// create lookup to know who to exclude from broadcast
		addresses := handler.peers.Addresses(peer.HasEntity(false, e.Hash))
		err = handler.net.Broadcast(e, addresses...)
		if err != nil {
			log.Error().Err(err).Msg("could not propagate entity")
			return
		}

		log.Debug().Msg("transaction processed")
	}
}

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
)

func (handler *Handler) processTransaction(wg *sync.WaitGroup, tx *types.Transaction) {
	defer wg.Done()

	// precompute the transaction hash
	tx.Hash = tx.GetHash()

	// configure logger
	log := handler.log.With().Str("component", "entity_transaction").Hex("hash", tx.Hash[:]).Logger()
	log.Debug().Msg("entity routine started")
	defer log.Debug().Msg("entity routine stopped")

	// check if we already know the transaction; if so, ignore it
	ok := handler.transactions.Has(tx.Hash)
	if ok {
		log.Debug().Msg("transaction already known")
		return
	}

	// check the validity of the transaction
	// TODO

	// add the transaction to the transaction pool
	err := handler.transactions.Add(tx)
	if err != nil {
		log.Error().Err(err).Msg("could not add transaction")
		return
	}

	handler.events.Transaction(tx.Hash)

	// create lookup to know who to exclude from broadcast
	addresses := handler.peers.Addresses(peer.HasEntity(false, tx.Hash))
	err = handler.net.Broadcast(tx, addresses...)
	if err != nil {
		log.Error().Err(err).Msg("could not propagate entity")
		return
	}

	log.Debug().Msg("transaction processed")
}

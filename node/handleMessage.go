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

func handleMessage(log zerolog.Logger, wg *sync.WaitGroup, handlers Handlers, net Network, state stateManager, pool poolManager, address string, message interface{}) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "processor").Str("address", address).Logger()
	log.Debug().Msg("processing routine started")
	defer log.Debug().Msg("processing routine stopped")

	// process the message according to type
	switch msg := message.(type) {

	case *types.Transaction:

		// TODO: validate the transaction

		// tag the peer for having seen the transaction
		id := msg.ID()
		state.Tag(address, id)

		// check if we already know the transaction; if so, ignore it
		ok := pool.Known(id)
		if ok {
			log.Debug().Msg("transaction already known")
			return
		}

		// add the transaction to the transaction pool
		err := pool.Add(msg)
		if err != nil {
			log.Error().Err(err).Msg("could not add transaction to pool")
			return
		}

		// handle the transaction for our blockchain state & propagation
		handlers.Entity(msg)

	case *Mempool:

		// find transactions in our memory pool the peer misses
		var inv [][]byte
		ids := pool.IDs()
		for _, id := range ids {
			if msg.Bloom.Test(id) {
				// TODO: figure out implications of false positives here
				state.Tag(address, id)
				continue
			}
			inv = append(inv, id)
		}

		// send the list of transaction IDs they do not have
		inventory := &Inventory{IDs: inv}
		err := net.Send(address, inventory)
		if err != nil {
			log.Error().Err(err).Msg("could not share inventory")
			return
		}

	case *Inventory:

		// create list of transactions that we are missing
		var req [][]byte
		for _, id := range msg.IDs {
			ok := pool.Known(id)
			if ok {
				continue
			}
			req = append(req, id)
		}

		// request the missing transactions from the peer
		// TODO: better to add into a pending queue
		request := &Request{IDs: req}
		err := net.Send(address, request)
		if err != nil {
			log.Error().Err(err).Msg("could not request transactions")
			return
		}

	case *Request:

		// for each requested ID
		for _, id := range msg.IDs {

			// check if we have the requested transaction
			tx, err := pool.Get(id)
			if err != nil {
				// TODO: somehow punish peer for requesting something we didn't announce
				log.Error().Err(err).Msg("requested transaction unknown")
				continue
			}

			// then send the peer the requested transaction
			err = net.Send(address, tx)
			if err != nil {
				log.Error().Err(err).Msg("could not send requested transaction")
				continue
			}
		}
	}
}

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
	"github.com/alvalor/alvalor-go/types"
)

func handleProcessing(log zerolog.Logger, wg *sync.WaitGroup, handlers handlerManager, pool poolManager, net networkManager, event network.Received) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "processor").Logger()
	log.Debug().Msg("processing routine started")
	defer log.Debug().Msg("processing routine stopped")

	// process the message according to type
	switch msg := event.Message.(type) {

	case *types.Transaction:
		ok := pool.Known(msg.ID())
		if ok {
			log.Debug().Msg("transaction already known")
			return
		}
		err := pool.Add(msg)
		if err != nil {
			log.Error().Err(err).Msg("could not add transaction to pool")
			return
		}
		handlers.Propagate(msg)

	case *Mempool:
		var inv [][]byte
		ids := pool.IDs()
		for _, id := range ids {
			if msg.Bloom.Test(id) {
				continue
			}
			inv = append(inv, id)
		}
		inventory := &Inventory{IDs: inv}
		err := net.Send(event.Address, inventory)
		if err != nil {
			log.Error().Err(err).Msg("could not share inventory")
			return
		}

	case *Inventory:
		var req [][]byte
		for _, id := range msg.IDs {
			ok := pool.Known(id)
			if ok {
				continue
			}
			req = append(req, id)
		}
		request := &Request{IDs: req}
		err := net.Send(event.Address, request)
		if err != nil {
			log.Error().Err(err).Msg("could not request transactions")
			return
		}

	case *Request:
		for _, id := range msg.IDs {
			tx, err := pool.Get(id)
			if err != nil {
				log.Error().Err(err).Msg("requested transaction unknown")
				continue
			}
			err = net.Send(event.Address, tx)
			if err != nil {
				log.Error().Err(err).Msg("could not send requested transaction")
				continue
			}
		}
	}
}

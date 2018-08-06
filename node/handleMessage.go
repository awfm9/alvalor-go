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

func handleMessage(log zerolog.Logger, wg *sync.WaitGroup, net Network, chain Blockchain, finder pathfinder, peers peerManager, pool poolManager, handlers Handlers, address string, message interface{}) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "processor").Str("address", address).Logger()
	log.Debug().Msg("processing routine started")
	defer log.Debug().Msg("processing routine stopped")

	// process the message according to type
	switch msg := message.(type) {

	case *Status:

		log = log.With().Str("msg_type", "status").Uint64("distance", msg.Distance).Logger()

		// if we are on a better path, we can ignore the status message
		path, distance := finder.Longest()
		if distance >= msg.Distance {
			log.Debug().Msg("not behind peer")
			return
		}

		// collect headers from the top of our longest path backwards
		// use increasing distance after first 8, finish with root (genesis)
		var locators []types.Hash
		index := 0
		step := 1
		for index < len(path)-1 {
			locators = append(locators, path[index])
			if len(locators) >= 8 {
				step *= 2
			}
			index += step
		}
		locators = append(locators, path[len(path)-1])

		// send synchronization message
		sync := &Sync{
			Locators: locators,
		}
		err := net.Send(address, sync)
		if err != nil {
			log.Error().Err(err).Msg("could not send sync message")
			return
		}

		log.Debug().Msg("processed status message")

	case *Sync:

		log = log.With().Str("msg_type", "sync").Int("num_locators", len(msg.Locators)).Logger()

		// create index of all locator hashes
		lookup := make(map[types.Hash]struct{})
		for _, locator := range msg.Locators {
			lookup[locator] = struct{}{}
		}

		// collect all header hashes on our best path until we run into a locator
		path, _ := finder.Longest()
		var hashes []types.Hash
		for _, hash := range path {
			_, ok := lookup[hash]
			if ok {
				break
			}
			hashes = append(hashes, hash)
		}

		// collect all the headers from our pathfinder
		// go in reverse order so we start with the oldest header first
		var headers []*types.Header
		for i := len(hashes) - 1; i >= 0; i-- {
			hash := hashes[i]
			header, err := finder.Header(hash)
			if err != nil {
				log.Error().Err(err).Hex("hash", hash[:]).Msg("could not retrieve header")
				return
			}
			headers = append(headers, header)
		}

		// send the partial path to our current distance to the other node
		p := Path{
			Headers: headers,
		}
		err := net.Send(address, p)
		if err != nil {
			log.Error().Err(err).Msg("could not send path")
			return
		}

		log.Debug().Msg("processed sync message")

	case *Path:

		log = log.With().Str("msg_type", "path").Int("num_headers", len(msg.Headers)).Logger()

		for _, header := range msg.Headers {
			handlers.Entity(header)
		}

		log.Debug().Msg("processed path message")

	case *types.Transaction:

		log = log.With().Str("msg_type", "transaction").Logger()

		handlers.Entity(msg)

		log.Debug().Msg("processed transaction message")

	case *Mempool:

		log = log.With().Str("msg_type", "mempool").Uint("num_cap", msg.Bloom.Cap()).Logger()

		// find transactions in our memory pool the peer misses
		var inv []types.Hash
		hashes := pool.Hashes()
		for _, hash := range hashes {
			if msg.Bloom.Test(hash[:]) {
				// TODO: figure out implications of false positives here
				peers.Tag(address, hash)
				continue
			}
			inv = append(inv, hash)
		}

		log = log.With().Int("num_inv", len(inv)).Logger()

		// send the list of transaction IDs they do not have
		inventory := &Inventory{Hashes: hashes}
		err := net.Send(address, inventory)
		if err != nil {
			log.Error().Err(err).Msg("could not share inventory")
			return
		}

		log.Debug().Msg("processed mempool message")

	case *Inventory:

		log = log.With().Int("num_inv", len(msg.Hashes)).Logger()

		// create list of transactions that we are missing
		var req []types.Hash
		for _, hash := range msg.Hashes {
			ok := pool.Known(hash)
			if ok {
				continue
			}
			req = append(req, hash)
		}

		log = log.With().Int("num_req", len(req)).Logger()

		// request the missing transactions from the peer
		handlers.RequestTransactions(req, address)

		log.Debug().Msg("processed inventory message, enqued request")

	case *Request:

		log = log.With().Int("num_req", len(msg.Hashes)).Logger()

		// collect each transaction that we have from the set of requested IDs
		var transactions []*types.Transaction
		for _, hash := range msg.Hashes {
			tx, err := pool.Get(hash)
			if err != nil {
				// TODO: somehow punish peer for requesting something we didn't announce
				log.Error().Err(err).Hex("hash", hash[:]).Msg("requested transaction unknown")
				continue
			}
			transactions = append(transactions, tx)
		}

		log = log.With().Int("num_txs", len(transactions)).Logger()

		// send the transactions in a batch
		batch := &Batch{Transactions: transactions}
		err := net.Send(address, batch)
		if err != nil {
			log.Error().Err(err).Msg("could not send transactions")
			return
		}

		log.Debug().Msg("processed request message")

	case *Batch:

		log = log.With().Int("num_txs", len(msg.Transactions)).Logger()

		for _, tx := range msg.Transactions {

			// TODO: validate transaction

			// tag the peer for having seen the transaction
			peers.Tag(address, tx.Hash)

			// handle the transaction for our blockchain state & propagation
			handlers.Entity(tx)
		}

		log.Debug().Msg("processed batch message")
	}
}

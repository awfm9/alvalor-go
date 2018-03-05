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
	"bytes"
	"encoding/hex"
	"sync"

	"github.com/rs/zerolog"

	"github.com/alvalor/alvalor-go/types"
)

func handleMessage(log zerolog.Logger, wg *sync.WaitGroup, handlers Handlers, net Network, chain chainManager, path pathManager, peers peerManager, pool poolManager, address string, message interface{}) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "processor").Str("address", address).Logger()
	log.Debug().Msg("processing routine started")
	defer log.Debug().Msg("processing routine stopped")

	// process the message according to type
	switch msg := message.(type) {

	case *Status:

		log.Info().Uint32("height", msg.Height).Msg("status message received")

		// TODO: should we save the height to the peer state?

		// get local best validated height from blockchain DB
		height, err := chain.Height()
		if err != nil {
			log.Error().Err(err).Msg("could not get chain height")
			return
		}

		// don't take any action if we are not behind the peer
		if msg.Height <= height {
			log.Debug().Uint32("height", height).Msg("chain height ahead of peer")
			return
		}

		// check if we are already synching the path to this unknown hash
		ok := path.Known(msg.Hash)
		if ok {
			log.Debug().Str("hash", hex.EncodeToString(msg.Hash)).Msg("path already synching")
			return
		}

		// get the best block hash from local blockchain DB
		bestBlock, err := chain.BestHash()
		if err != nil {
			log.Error().Err(err).Msg("could not get best block hash")
			return
		}

		// include the best header hash if it's different from the best validated block
		var locators [][]byte
		bestHeader := path.BestHash()
		if !bytes.Equal(bestHeader, bestBlock) {
			locators = append(locators, bestHeader)
		}

		// decide which block locator hashes to include by height
		var heights []uint32
		for h, step := height, uint32(1); h > 0; h -= step {
			if len(heights) >= 8 {
				step *= 2
			}
			heights = append(heights, h)
		}
		heights = append(heights, 0)

		// retrieve the hashes from the blockchain database
		for _, height := range heights {
			hash, err := chain.HashByHeight(height)
			if err != nil {
				log.Error().Err(err).Msg("could not get hash by index")
				return
			}
			locators = append(locators, hash)
		}

		// create the synchronization request & send
		sync := &Sync{
			Locators: locators,
		}
		err = net.Send(address, sync)
		if err != nil {
			log.Error().Err(err).Msg("could not send synchronization")
			return
		}

	case *types.Transaction:

		log.Info().Str("id", hex.EncodeToString(msg.ID())).Msg("transaction message received")

		// TODO: validate the transaction

		// tag the peer for having seen the transaction
		peers.Tag(address, msg.ID())

		// handle the transaction for our blockchain state & propagation
		handlers.Entity(msg)

	case *Mempool:

		log.Info().Uint("cap", msg.Bloom.Cap()).Msg("bloom message received")

		// find transactions in our memory pool the peer misses
		var inv [][]byte
		ids := pool.IDs()
		for _, id := range ids {
			if msg.Bloom.Test(id) {
				// TODO: figure out implications of false positives here
				peers.Tag(address, id)
				continue
			}
			inv = append(inv, id)
		}

		log.Info().Int("num_ids", len(inv)).Msg("received mempool message")

		// send the list of transaction IDs they do not have
		inventory := &Inventory{IDs: inv}
		err := net.Send(address, inventory)
		if err != nil {
			log.Error().Err(err).Msg("could not share inventory")
			return
		}

	case *Inventory:

		log.Info().Int("num_ids", len(msg.IDs)).Msg("received inventory message")

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

		log.Info().Int("num_ids", len(msg.IDs)).Msg("received request message")

		// collect each transaction that we have from the set of requested IDs
		var transactions []*types.Transaction
		for _, id := range msg.IDs {
			tx, err := pool.Get(id)
			if err != nil {
				// TODO: somehow punish peer for requesting something we didn't announce
				log.Error().Err(err).Msg("requested transaction unknown")
				continue
			}
			transactions = append(transactions, tx)
		}

		// send the transactions in a batch
		batch := &Batch{Transactions: transactions}
		err := net.Send(address, batch)
		if err != nil {
			log.Error().Err(err).Msg("could not send transactions")
			return
		}

	case *Batch:

		log.Info().Int("num_txs", len(msg.Transactions)).Msg("received batch message")

		for _, tx := range msg.Transactions {

			// TODO: validate transaction

			// tag the peer for having seen the transaction
			peers.Tag(address, tx.ID())

			// handle the transaction for our blockchain state & propagation
			handlers.Entity(tx)
		}
	}
}

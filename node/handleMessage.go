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

func handleMessage(log zerolog.Logger, wg *sync.WaitGroup, net Network, chain Blockchain, path Path, peers peerManager, pool poolManager, handlers Handlers, address string, message interface{}) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "processor").Str("address", address).Logger()
	log.Debug().Msg("processing routine started")
	defer log.Debug().Msg("processing routine stopped")

	// process the message according to type
	switch msg := message.(type) {

	case *Status:

		log = log.With().Uint32("height", msg.Height).Str("hash", hex.EncodeToString(msg.Hash)).Logger()

		// don't take any action if we are not behind the peer
		height := chain.Height()
		if msg.Height <= chain.Height() {
			log.Debug().Msg("already synced with peer")
			return
		}

		// check if we are already synching the path to this unstored block
		ok := path.Has(msg.Hash)
		if ok {
			log.Debug().Msg("already syncing with peer")
			return
		}

		// // add the latest synching header to our locator hashes if it's different from chain state
		var locators [][]byte
		// bestBlock := chain.Current().Hash()
		// bestHeader := path.BestHash()
		// if !bytes.Equal(bestHeader, bestBlock) {
		// 	locators = append(locators, bestHeader)
		// }

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
				log.Error().Err(err).Uint32("height", height).Msg("could not get hash by height")
				return
			}
			locators = append(locators, hash)
		}

		// create the synchronization request & send
		sync := &Sync{
			Locators: locators,
		}
		err := net.Send(address, sync)
		if err != nil {
			log.Error().Err(err).Msg("could not send synchronization")
			return
		}

		log.Debug().Msg("replied with synchronization message")

	case *Sync:

		locators := make([]string, 0, len(msg.Locators))
		for _, locator := range msg.Locators {
			locators = append(locators, hex.EncodeToString(locator))
		}

		log = log.With().Strs("locators", locators).Logger()

		// try finding a locator hash in our best path
		var start uint32
	Outer:
		for _, locator := range msg.Locators {
			header := &chain.Current().Header
			for {
				hash := header.Hash()
				if bytes.Equal(hash, locator) {
					var err error
					start, err = chain.HeightByHash(header.Hash())
					if err != nil {
						log.Error().Err(err).Msg("could not get height by hash")
						return
					}
					break Outer
				}
				parent := header.Parent
				if bytes.Equal(parent, bytes.Repeat([]byte{0}, 32)) {
					break
				}
				var err error
				header, err = chain.HeaderByHash(parent)
				if err != nil {
					log.Error().Err(err).Str("hash", hex.EncodeToString(parent)).Msg("could not header by hash")
					return
				}
			}
		}

		// return all headers from the found start to the top
		var headers []*types.Header
		for height := start + 1; height <= chain.Height(); height++ {
			header, err := chain.HeaderByHeight(height)
			if err != nil {
				log.Error().Err(err).Uint32("height", height).Msg("could not get header by height")
				return
			}
			headers = append(headers, header)
		}

		// send each header
		for _, header := range headers {
			err := net.Send(address, header)
			if err != nil {
				log.Error().Err(err).Msg("could not send header")
				return
			}
		}

		log.Debug().Msg("processed synchronization message")

	case *types.Header:

		hash := msg.Hash()
		log = log.With().Str("msg_type", "header").Str("hash", hex.EncodeToString(hash)).Logger()

		// check if we already stored the header
		_, err := chain.HeaderByHash(hash)
		if err == nil {
			log.Debug().Msg("header already known")
			return
		}

		log.Debug().Msg("processed header message")

	case *types.Transaction:

		hash := msg.Hash()
		log = log.With().Str("msg_type", "transaction").Str("hash", hex.EncodeToString(hash)).Logger()

		// check if we already know the transaction
		_, err := chain.TransactionByHash(hash)
		if err == nil {
			log.Debug().Msg("transaction already known")
			return
		}

		// TODO: validate the transaction

		// tag the peer for having seen the transaction
		peers.Tag(address, hash)

		// handle the transaction for our blockchain state & propagation
		handlers.Entity(msg)

		log.Debug().Msg("processed transaction message")

	case *Mempool:

		log = log.With().Str("msg_type", "mempool").Uint("num_cap", msg.Bloom.Cap()).Logger()

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

		log = log.With().Int("num_inv", len(inv)).Logger()

		// send the list of transaction IDs they do not have
		inventory := &Inventory{IDs: inv}
		err := net.Send(address, inventory)
		if err != nil {
			log.Error().Err(err).Msg("could not share inventory")
			return
		}

		log.Debug().Msg("processed mempool message")

	case *Inventory:

		log = log.With().Int("num_inv", len(msg.IDs)).Logger()

		// create list of transactions that we are missing
		var req [][]byte
		for _, id := range msg.IDs {
			ok := pool.Known(id)
			if ok {
				continue
			}
			req = append(req, id)
		}

		log = log.With().Int("num_req", len(req)).Logger()

		// request the missing transactions from the peer
		// TODO: better to add into a pending queue
		request := &Request{IDs: req}
		err := net.Send(address, request)
		if err != nil {
			log.Error().Err(err).Msg("could not request transactions")
			return
		}

		log.Debug().Msg("processed inventory message")

	case *Request:

		log = log.With().Int("num_req", len(msg.IDs)).Logger()

		// collect each transaction that we have from the set of requested IDs
		var transactions []*types.Transaction
		for _, hash := range msg.IDs {
			tx, err := pool.Get(hash)
			if err != nil {
				// TODO: somehow punish peer for requesting something we didn't announce
				log.Error().Err(err).Str("hash", hex.EncodeToString(hash)).Msg("requested transaction unknown")
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
			peers.Tag(address, tx.Hash())

			// handle the transaction for our blockchain state & propagation
			handlers.Entity(tx)
		}

		log.Debug().Msg("processed batch message")
	}
}

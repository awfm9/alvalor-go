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

func handleMessage(log zerolog.Logger, wg *sync.WaitGroup, net Network, chain Blockchain, finder Finder, peers peerManager, pool poolManager, handlers Handlers, address string, message interface{}) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "processor").Str("address", address).Logger()
	log.Debug().Msg("processing routine started")
	defer log.Debug().Msg("processing routine stopped")

	// process the message according to type
	switch msg := message.(type) {

	case *Status:

		log = log.With().Str("msg_type", "status").Uint32("height", msg.Height).Hex("hash", msg.Hash[:]).Logger()

		// don't take any action if we are not behind the peer
		height := chain.Height()
		if msg.Height <= chain.Height() {
			log.Debug().Msg("not behind peer height")
			return
		}

		// check if we are already synching the path to this unstored block
		ok := finder.Has(msg.Hash)
		if ok {
			log.Debug().Msg("already syncing potential path")
			return
		}

		// // add the latest synching header to our locator hashes if it's different from chain state
		var locators []types.Hash
		// TODO: rethink & implement
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

		log = log.With().Int("locators", len(locators)).Logger()

		// create the synchronization request & send
		sync := &Sync{
			Locators: locators,
		}
		err := net.Send(address, sync)
		if err != nil {
			log.Error().Err(err).Msg("could not send synchronization")
			return
		}

		log.Debug().Msg("processed status message")

	case *Sync:

		log = log.With().Str("msg_type", "sync").Logger()

		// try finding a locator hash in our best path
		var common uint32
	Outer:
		for _, locator := range msg.Locators {
			header := chain.Header()
			for {
				hash := header.Hash()
				if hash == locator {
					log = log.With().Hex("hash", hash[:]).Logger()
					var err error
					common, err = chain.HeightByHash(header.Hash())
					if err != nil {
						log.Error().Err(err).Msg("could not get height by hash")
						return
					}
					log = log.With().Uint32("common_height", common).Logger()
					break Outer
				}
				parent := header.Parent
				if parent == types.ZeroHash {
					break
				}
				var err error
				header, err = chain.HeaderByHash(parent)
				if err != nil {
					log.Error().Err(err).Hex("hash", parent[:]).Msg("could not get header by hash")
					return
				}
			}
		}

		// return all headers from the found start to the top
		var headers []*types.Header
		for height := common + 1; height <= chain.Height(); height++ {
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

		log = log.With().Str("msg_type", "header").Hex("hash", hash[:]).Hex("parent", msg.Parent[:]).Logger()

		// check if we already stored the header
		_, err := chain.HeaderByHash(msg.Hash())
		if err == nil {
			log.Debug().Msg("header already known")
			return
		}

		// check if we already process the header
		ok := finder.Has(msg.Hash())
		if ok {
			log.Debug().Msg("header already processing")
			return
		}

		// add the header to the path finder
		err = finder.Add(msg.Hash(), msg.Parent)
		if err != nil {
			log.Error().Err(err).Msg("could not add header to path")
			return
		}

		// collect all information needed to complete this path
		path := finder.Path()
		handlers.Collect(path)

		log.Debug().Msg("processed header message")

	case *types.Transaction:

		hash := msg.Hash()

		log = log.With().Str("msg_type", "transaction").Hex("hash", hash[:]).Logger()

		// check if we already know the transaction
		_, err := chain.TransactionByHash(msg.Hash())
		if err == nil {
			log.Debug().Msg("transaction already known")
			return
		}

		// TODO: validate the transaction

		// tag the peer for having seen the transaction
		peers.Tag(address, msg.Hash())

		// handle the transaction for our blockchain state & propagation
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
		// TODO: better to add into a pending queue
		request := &Request{Hashes: req}
		err := net.Send(address, request)
		if err != nil {
			log.Error().Err(err).Msg("could not request transactions")
			return
		}

		log.Debug().Msg("processed inventory message")

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
			peers.Tag(address, tx.Hash())

			// handle the transaction for our blockchain state & propagation
			handlers.Entity(tx)
		}

		log.Debug().Msg("processed batch message")
	}
}

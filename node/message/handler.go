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

package message

import (
	"sync"

	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
)

// Handler represents the handler for messages from the network stack.
type Handler struct {
	log          zerolog.Logger
	net          Network
	paths        Paths
	downloads    Downloads
	headers      Headers
	inventories  Inventories
	transactions Transactions
	peers        Peers
	entity       Entity
}

// Process processes a message from the network.
func (handler *Handler) Process(wg *sync.WaitGroup, address string, message interface{}) {
	wg.Add(1)
	switch msg := message.(type) {
	case *Status:
		go handler.processStatus(wg, address, msg)
	case *Sync:
		go handler.processSync(wg, address, msg)
	case *Path:
		go handler.processPath(wg, address, msg)
	case *GetInv:
		go handler.processGetInv(wg, address, msg)
	case *GetTx:
		go handler.processGetTx(wg, address, msg)
	case *types.Inventory:
		go handler.processInventory(wg, address, msg)
	case *types.Transaction:
		go handler.processTransaction(wg, address, msg)
	}
}

func (handler *Handler) process(wg *sync.WaitGroup, address string, message interface{}) {
	defer wg.Done()

	// configure logger
	log := handler.log.With().Str("component", "message").Str("address", address).Logger()
	log.Debug().Msg("message routine started")
	defer log.Debug().Msg("message routine stopped")

	// process the message according to type
	switch msg := message.(type) {

	// The Sync message is a request for block headers. It contains a number
	// of locator hashes that allows the receiving peer to search a common
	// block header hash on his best path. The receiving peer will then send a
	// a Path message with the missing headers. Ideally, they are sent in
	// chronological order, from oldest to newest, to speed up processing.
	case *Sync:

		log = log.With().Str("msg_type", "sync").Int("num_locators", len(msg.Locators)).Logger()

		// create lookup table of locator hashes
		lookup := make(map[types.Hash]struct{})
		for _, locator := range msg.Locators {
			lookup[locator] = struct{}{}
		}

		// collect all header hashes on our best path until we run into a locator
		path, _ := handler.headers.Path()
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
		var hdrs []*types.Header
		for i := len(hashes) - 1; i >= 0; i-- {
			hash := hashes[i]
			header, err := handler.headers.Get(hash)
			if err != nil {
				log.Error().Err(err).Hex("hash", hash[:]).Msg("could not retrieve header")
				return
			}
			hdrs = append(hdrs, header)
		}

		// send the partial path to our best distance to the other node
		p := &Path{
			Headers: hdrs,
		}
		err := handler.net.Send(address, p)
		if err != nil {
			log.Error().Err(err).Msg("could not send path")
			return
		}

		log.Debug().Msg("processed sync message")

	// The Path message is a reply to the Sync message, which contains the missing
	// block headers on the best path, as identified by the locator hashes. They
	// should be ordered by chronological order, from oldest to newest, in order
	// to allow the most efficient construction of the best path.
	case *Path:

		log = log.With().Str("msg_type", "path").Int("num_headers", len(msg.Headers)).Logger()

		for _, header := range msg.Headers {
			handler.entity.Process(wg, header)
		}

		log.Debug().Msg("processed path message")

	// The GetInv is a message sent by peers who want to download the given
	// block inventory from us. If we have it, we send it to them.
	// TODO: reply with not available if we don't have it
	case *GetInv:

		log = log.With().Str("msg_type", "get_inv").Hex("hash", msg.Hash[:]).Logger()

		// try to get the inventory
		inv, err := handler.inventories.Get(msg.Hash)
		if err != nil {
			log.Error().Err(err).Msg("could not get inventory")
			return
		}

		// try to send the inventory
		err = handler.net.Send(address, inv)
		if err != nil {
			log.Error().Err(err).Msg("could not send inventory")
			return
		}

		log.Debug().Msg("processed get_inv message")

	// The GetTx is a message sent by peers who want to download the given
	// transaction from us. If we have it, we send it to them.
	// TODO: reply with not available if we don't have it
	case *GetTx:

		log = log.With().Str("msg_type", "get_tx").Hex("hash", msg.Hash[:]).Logger()

		// try to get the inventory
		tx, err := handler.transactions.Get(msg.Hash)
		if err != nil {
			log.Error().Err(err).Msg("could not get transaction")
			return
		}

		// try to send the inventory
		err = handler.net.Send(address, tx)
		if err != nil {
			log.Error().Err(err).Msg("could not send transaction")
			return
		}

		log.Debug().Msg("processed get_tx message")

	// The inventory is a the template of how to reconstruct a block from messages
	// and is used to download all necessary messages to fully validate a block.
	case *types.Inventory:

		log = log.With().Str("msg_type", "inventory").Hex("hash", msg.Hash[:]).Int("num_hashes", len(msg.Hashes)).Logger()

		// cancel any pending download retries for this inventory
		handler.downloads.Cancel(msg.Hash)

		// mark the inventory as received for the respective peer
		handler.peers.Received(address, msg.Hash)

		// store the new inventory in our database
		err := handler.inventories.Add(msg)
		if err != nil {
			log.Error().Err(err).Msg("could not store received inventory")
			return
		}

		// signal the new inventory to the tracker to start pending tx downloads
		err = handler.paths.Signal(msg.Hash)
		if err != nil {
			log.Error().Err(err).Msg("could not signal inventory")
			return
		}

		log.Debug().Msg("processed inventory message")

	// The transaction message is a message containing a transaction.
	case *types.Transaction:

		log = log.With().Str("msg_type", "transaction").Hex("hash", msg.Hash[:]).Logger()

		// cancel any pending download retries for this transaction
		handler.downloads.Cancel(msg.Hash)

		// mark the inventory download as completed for the respective peer
		handler.peers.Received(address, msg.Hash)

		// handle the transaction entity
		handler.entity.Process(wg, msg)

		log.Debug().Msg("processed transaction message")
	}
}

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

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/alvalor/alvalor-go/types"

	"github.com/alvalor/alvalor-go/node/message"
	"github.com/alvalor/alvalor-go/node/repo"
)

func handleMessage(log zerolog.Logger, wg *sync.WaitGroup, net Network, paths Paths, downloads Downloads, headers Headers, inventories Inventories, transactions Transactions, peers Peers, handlers Handlers, address string, input interface{}) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "processor").Str("address", address).Logger()
	log.Debug().Msg("processing routine started")
	defer log.Debug().Msg("processing routine stopped")

	// process the message according to type
	switch msg := input.(type) {

	// The Status message is a handshake sent by both peers on a new connection.
	// It contains the distance of their best path and helps each peer to
	// determine whether they should request missing headers from the other. If a
	// peer is behind, it should send a Sync message with a number of locator
	// hashes of block headers, to request the missing headers from the peer who
	// is ahead.
	case *message.Status:

		log = log.With().Str("msg_type", "status").Uint64("distance", msg.Distance).Logger()

		// if we are on a better path, we can ignore the status message
		path, distance := headers.Path()
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
		sync := &message.Sync{
			Locators: locators,
		}
		err := net.Send(address, sync)
		if err != nil {
			log.Error().Err(err).Msg("could not send sync message")
			return
		}

		log.Debug().Msg("processed status message")

	// The Sync message is a request for block headers. It contains a number
	// of locator hashes that allows the receiving peer to search a common
	// block header hash on his best path. The receiving peer will then send a
	// a Path message with the missing headers. Ideally, they are sent in
	// chronological order, from oldest to newest, to speed up processing.
	case *message.Sync:

		log = log.With().Str("msg_type", "sync").Int("num_locators", len(msg.Locators)).Logger()

		// create lookup table of locator hashes
		lookup := make(map[types.Hash]struct{})
		for _, locator := range msg.Locators {
			lookup[locator] = struct{}{}
		}

		// collect all header hashes on our best path until we run into a locator
		path, _ := headers.Path()
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
			header, err := headers.Get(hash)
			if err != nil {
				log.Error().Err(err).Hex("hash", hash[:]).Msg("could not retrieve header")
				return
			}
			hdrs = append(hdrs, header)
		}

		// send the partial path to our best distance to the other node
		p := &message.Path{
			Headers: hdrs,
		}
		err := net.Send(address, p)
		if err != nil {
			log.Error().Err(err).Msg("could not send path")
			return
		}

		log.Debug().Msg("processed sync message")

	// The Path message is a reply to the Sync message, which contains the missing
	// block headers on the best path, as identified by the locator hashes. They
	// should be ordered by chronological order, from oldest to newest, in order
	// to allow the most efficient construction of the best path.
	case *message.Path:

		log = log.With().Str("msg_type", "path").Int("num_headers", len(msg.Headers)).Logger()

		for _, header := range msg.Headers {
			handlers.entity(header)
		}

		log.Debug().Msg("processed path message")

	// The Request message is a request sent by peers who want to download the
	// entity with the given hash. It can be used for inventories, transactions
	// or blocks.
	case *message.Request:

		log = log.With().Str("msg_type", "confirm").Hex("hash", msg.Hash[:]).Logger()

		// try to respond with an inventory
		err := respondInventory(net, address, msg.Hash, inventories)
		if err == nil {
			log.Debug().Msg("processed request message (inventory)")
			return
		}
		if errors.Cause(err) != repo.ErrNotFound {
			log.Error().Err(err).Msg("could not check inventory store")
			return
		}

		// try to respond with a transaction
		err = respondTransaction(net, address, msg.Hash, transactions)
		if err == nil {
			log.Debug().Msg("processed request message (transaction)")
			return
		}
		if errors.Cause(err) != repo.ErrNotFound {
			log.Error().Err(err).Msg("could not check transaction pool")
			return
		}

		log.Debug().Msg("processed request message (entity not found)")

	case *types.Inventory:

		log = log.With().Str("msg_type", "inventory").Hex("hash", msg.Hash[:]).Int("num_hashes", len(msg.Hashes)).Logger()

		// cancel any pending download retries for this inventory
		downloads.Cancel(msg.Hash)

		// mark the inventory as received for the respective peer
		peers.Received(address, msg.Hash)

		// store the new inventory in our database
		err := inventories.Add(msg)
		if err != nil {
			log.Error().Err(err).Msg("could not store received inventory")
			return
		}

		// signal the new inventory to the tracker to start pending tx downloads
		err = paths.Signal(msg.Hash)
		if err != nil {
			log.Error().Err(err).Msg("could not signal inventory")
			return
		}

		log.Debug().Msg("processed inventory message")

	case *types.Transaction:

		log = log.With().Str("msg_type", "transaction").Hex("hash", msg.Hash[:]).Logger()

		// cancel any pending download retries for this transaction
		downloads.Cancel(msg.Hash)

		// mark the inventory download as completed for the respective peer
		peers.Received(address, msg.Hash)

		// handle the transaction entity
		handlers.entity(msg)

		log.Debug().Msg("processed transaction message")
	}
}

func respondInventory(net Network, address string, hash types.Hash, inventories Inventories) error {
	inv, err := inventories.Get(hash)
	if errors.Cause(err) == repo.ErrNotFound {
		return errors.Wrap(err, "could not find inventory")
	}
	err = net.Send(address, inv)
	if err != nil {
		return errors.Wrap(err, "could not send inventory")
	}
	return nil
}

func respondTransaction(net Network, address string, hash types.Hash, transactions Transactions) error {
	tx, err := transactions.Get(hash)
	if errors.Cause(err) == repo.ErrNotFound {
		return errors.Wrap(err, "could not find transaction")
	}
	err = net.Send(address, tx)
	if err != nil {
		return errors.Wrap(err, "could not send transaction")
	}
	return nil
}

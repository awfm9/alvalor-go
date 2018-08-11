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

func handleMessage(log zerolog.Logger, wg *sync.WaitGroup, net Network, finder pathfinder, chain blockchain, handlers Handlers, address string, message interface{}) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "processor").Str("address", address).Logger()
	log.Debug().Msg("processing routine started")
	defer log.Debug().Msg("processing routine stopped")

	// process the message according to type
	switch msg := message.(type) {

	// The Status message is handshake sent by both peers on a new connection.
	// It contains the distance of their best path and helps each peer to
	// determine whether they should request missing headers from the other. If a
	// peer is behind, it should send a Sync message with a number of locator
	// hashes of block headers, to request the missing headers from the peer who
	// is ahead.
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

	// The Sync message is a request for block headers. It contains a number
	// of locator hashes that allows the receiving peer to find the last common
	// block header with the requesting peer on the best path. The receiving peer
	// should then send a Path message with the missing headers in chronological
	// order, from oldest to newest.
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

	// The Path message is a reply to the Sync message, which contains the missing
	// block headers on the best path, as identified by the locator hashes. They
	// should be ordered by chronological order, from oldest to newest, in order
	// to allow the most efficient construction of the best path.
	case *Path:

		log = log.With().Str("msg_type", "path").Int("num_headers", len(msg.Headers)).Logger()

		for _, header := range msg.Headers {
			handlers.Entity(header)
		}

		log.Debug().Msg("processed path message")

	// The Confirm message is a request sent by peers who want to reconstruct a
	// certain block by downloading its transactions. It contains the hash of
	// the block, as well as a list of hashes of all the transactions. They can
	// then be queued for parralellized downloading.
	case *Confirm:

		log = log.With().Str("msg_type", "confirm").Hex("hash", msg.Hash[:]).Logger()

		// retrieve the hashes from the block manager
		var hashes []types.Hash
		hashes, err := chain.Inventory(msg.Hash)
		if err != nil {
			log.Error().Err(err).Msg("could not retrieve block inventory")
			return
		}

		// send the inventory message to the peer
		inv := &Inventory{
			Hash:   msg.Hash,
			Hashes: hashes,
		}
		err = net.Send(address, inv)
		if err != nil {
			log.Error().Err(err).Msg("could not send inventory message")
			return
		}

		log.Debug().Msg("processed confirm message")

	case *Inventory:

		log = log.With().Str("msg_type", "batch").Hex("hash", msg.Hash[:]).Int("num_hashes", len(msg.Hashes)).Logger()

		log.Debug().Msg("processed batch message")
	}
}

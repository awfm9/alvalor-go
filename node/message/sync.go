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
)

// The Sync message is a request for block headers. It contains a number
// of locator hashes that allows the receiving peer to search a common
// block header hash on his best path. The receiving peer will then send a
// a Path message with the missing headers. Ideally, they are sent in
// chronological order, from oldest to newest, to speed up processing.
func (handler *Handler) processSync(wg *sync.WaitGroup, address string, sync *Sync) {
	defer wg.Done()

	// configure logger
	with := handler.log.With()
	with.Str("component", "message")
	with.Str("message_type", "sync")
	with.Str("address", address)
	with.Int("num_locators", len(sync.Locators))
	log := with.Logger()

	// wrap routine in start and stop messages
	log.Debug().Msg("routine started")
	defer log.Debug().Msg("routine stopped")

	// create lookup table of locator hashes
	lookup := make(map[types.Hash]struct{})
	for _, locator := range sync.Locators {
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
}

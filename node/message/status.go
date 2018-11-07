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

func (handler *Handler) processStatus(wg *sync.WaitGroup, address string, status *Status) {
	defer wg.Done()

	// configure logger
	with := handler.log.With()
	with.Str("component", "message")
	with.Str("message_type", "status")
	with.Str("address", address)
	with.Uint64("distance", status.Distance)
	log := with.Logger()

	// wrap routine in start and stop messages
	log.Debug().Msg("routine started")
	defer log.Debug().Msg("routine stopped")

	// The Status message is a handshake sent by both peers on a new connection.
	// It contains the distance of their best path and helps each peer to
	// determine whether they should request missing headers from the other. If a
	// peer is behind, it should send a Sync message with a number of locator
	// hashes of block headers, to request the missing headers from the peer who
	// is ahead.

	// if we are on a better path, we can ignore the status message
	path, distance := handler.headers.Path()
	if distance >= status.Distance {
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
	err := handler.net.Send(address, sync)
	if err != nil {
		log.Error().Err(err).Msg("could not send sync message")
		return
	}

	log.Debug().Msg("processed status message")
}

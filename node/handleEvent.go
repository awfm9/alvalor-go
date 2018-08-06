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

func handleEvent(log zerolog.Logger, wg *sync.WaitGroup, net Network, finder pathfinder, peers peerManager, handlers Handlers, event interface{}) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "event").Logger()
	log.Debug().Msg("event routine started")
	defer log.Debug().Msg("event routine stopped")

	switch e := event.(type) {

	case network.Connected:

		peers.Active(e.Address)

		// get list of hashes from our best path and the current distance
		path, distance := finder.Longest()

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

		// send our current distance and locator hashes
		sync := &Sync{
			Distance: distance,
			Locators: locators,
		}
		err := net.Send(e.Address, sync)
		if err != nil {
			log.Error().Err(err).Msg("could not send status message")
			return
		}

	case network.Disconnected:
		peers.Inactive(e.Address)

	case network.Received:
		handlers.Message(e.Address, e.Message)
	}
}

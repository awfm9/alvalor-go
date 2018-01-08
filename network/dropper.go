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

package network

import (
	"math/rand"
	"sync"

	"github.com/rs/zerolog"
)

// Dropper are the dependencies dropping routines need.
type Dropper interface {
	PeerCount() uint
	GetAddresses() []string
	DropPeer(address string) error
}

func handleDropping(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, mgr Dropper, book *Book) {
	defer wg.Done()

	// extract desired configuration parameters
	var (
		maxPeers = cfg.maxPeers
	)

	// configure logger and add start/stop messages
	log = log.With().Str("component", "dropper").Logger()
	log.Info().Msg("dropping routine started")
	defer log.Info().Msg("dropping routine stopped")

	// if we don't have too many peers, abort
	if mgr.PeerCount() <= maxPeers {
		log.Debug().Msg("valid number of peers")
		return
	}

	// try to get addresses of peers available to drop
	addresses := mgr.GetAddresses()
	if len(addresses) == 0 {
		log.Debug().Msg("not connected to any peers")
		return
	}

	// select a random peer and drop it
	address := addresses[rand.Int()%len(addresses)]
	err := mgr.DropPeer(address)
	if err != nil {
		log.Error().Str("address", address).Err(err).Msg("could not drop peer")
		return
	}

	// notify the address manager that we dropped the peer
	book.Dropped(address)
}

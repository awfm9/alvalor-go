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
	"time"

	"github.com/rs/zerolog"
)

func handleDropping(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, peers peerManager, stop <-chan struct{}) {
	defer wg.Done()

	// extract desired configuration parameters
	var (
		interval = cfg.interval
		maxPeers = cfg.maxPeers
	)

	// configure logger and add start/stop messages
	log = log.With().Str("component", "dropper").Logger()
	log.Debug().Msg("dropping routine started")
	defer log.Debug().Msg("dropping routine stopped")

	// each tick, check if we have too many peers and if yes, drop one
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-stop:
			ticker.Stop()
			return
		case <-ticker.C:
		}
		if peers.Count() <= maxPeers {
			continue
		}
		addresses := peers.Addresses()
		address := addresses[rand.Int()%len(addresses)]
		err := peers.Drop(address)
		if err != nil {
			log.Error().Err(err).Msg("could not drop peer")
			continue
		}
	}
}

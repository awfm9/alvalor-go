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
	"crypto/sha256"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

func handleDialing(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, peers peerManager, pending pendingManager, book addressManager, rep reputationManager, handlers handlerManager, stop <-chan struct{}) {
	defer wg.Done()

	// extract needed configuration parameters
	var (
		address  = cfg.address
		interval = cfg.interval
		minPeers = cfg.minPeers
	)

	// configure logger and add start/stop messages
	log = log.With().Str("component", "dialer").Logger()
	log.Debug().Msg("dialing routine started")
	defer log.Debug().Msg("dialing routine stopped")

	// on each tick, check if we are below minimum peers and should have free
	// connection slots, then start a new dialer
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-stop:
			ticker.Stop()
			return
		case <-ticker.C:
		}
		peerCount := peers.Count()
		if peerCount >= minPeers {
			continue
		}
		sample := book.Sample(1,
			isNot([]string{address}),
			isNot(pending.Addresses()),
			isNot(peers.Addresses()),
			isScoreAbove(rep, -5),
			isFailBefore(rep, time.Now().Add(-15*time.Minute)),
			byScore(rep),
			byIPHash(sha256.New()),
		)
		if len(sample) == 0 {
			log.Debug().Msg("could not get address to connect")
			handlers.Discoverer()
			continue
		}
		handlers.Connector(sample[0])
	}
}

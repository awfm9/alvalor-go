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
	"sync"
	"time"

	"github.com/rs/zerolog"
)

func handleServing(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, peers peerManager, handlers handlerManager, stop <-chan struct{}) {
	defer wg.Done()

	// extract the configuration parameters we are interested in
	var (
		listen   = cfg.listen
		interval = cfg.interval
		maxPeers = cfg.maxPeers
	)

	// configure the logger for the component with start/stop messages
	log = log.With().Str("component", "server").Logger()
	log.Debug().Msg("serving routine started")
	defer log.Debug().Msg("serving routine stopped")

	// each time we tick, check if we should enable or disable the accepting of
	// connections
	var running bool
	var done chan struct{}
	ticker := time.NewTicker(interval)
Loop:
	for {
		select {
		case <-stop:
			break Loop
		case <-ticker.C:
		}
		peerCount := peers.Count()
		if peerCount < maxPeers && !running && listen {
			done = make(chan struct{})
			handlers.Listener()
			running = true
		} else if peerCount >= maxPeers && running {
			close(done)
			running = false
		}
	}

	// after the stop signal is received, we just need to stop listening if it's
	// currently on, as the external waitgroup will handle the rest, we don't
	// need to wait here
	ticker.Stop()
	if running {
		close(done)
	}
}

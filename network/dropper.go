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

// Dropper are the dependencies dropping routines need.
type Dropper interface {
	PeerCount() uint
	DropPeer() error
}

func handleDropping(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, mgr Dropper, stop <-chan struct{}) {
	defer wg.Done()
	log = log.With().Str("component", "dropper").Logger()
	log.Info().Msg("dropping routine started")
	defer log.Info().Msg("dropping routine stopped")
	ticker := time.NewTicker(cfg.interval)
	for {
		select {
		case <-stop:
			ticker.Stop()
			return
		case <-ticker.C:
		}
		if mgr.PeerCount() > cfg.maxPeers {
			err := mgr.DropPeer()
			if err != nil {
				log.Error().Err(err).Msg("could not drop peer")
				continue
			}
		}
	}
}

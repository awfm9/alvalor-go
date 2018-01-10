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

type dropperInfos interface {
	PeerCount() uint
}

type dropperActions interface {
	DropRandomPeer() (string, error)
}

type dropperEvents interface {
	Dropped(address string)
}

func handleDropping(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, infos dropperInfos, actions dropperActions, events dropperEvents, stop <-chan struct{}) {
	defer wg.Done()

	// extract desired configuration parameters
	var (
		interval = cfg.interval
		maxPeers = cfg.maxPeers
	)

	// configure logger and add start/stop messages
	log = log.With().Str("component", "dropper").Logger()
	log.Info().Msg("dropping routine started")
	defer log.Info().Msg("dropping routine stopped")

	// each tick, check if we have too many peers and if yes, drop one
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-stop:
			ticker.Stop()
			return
		case <-ticker.C:
		}
		if infos.PeerCount() <= maxPeers {
			continue
		}
		address, err := actions.DropRandomPeer()
		if err != nil {
			log.Error().Err(err).Msg("could not drop peer")
			continue
		}
		events.Dropped(address)
	}
}

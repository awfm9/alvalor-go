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

func handleDropping(log zerolog.Logger, wg *sync.WaitGroup, ticker <-chan time.Time, maxPeers uint, peerCount uintFunc, dropPeer errorFunc) {
	defer wg.Done()
	log = log.With().Str("component", "dropper").Logger()
	log.Info().Msg("peer dropping routine started")
	defer log.Info().Msg("peer dropping routine stopped")
	for _ = range ticker {
		numPeers := peerCount()
		if numPeers > maxPeers {
			err := dropPeer()
			if err != nil {
				log.Info().Err(err).Msg("could not drop peer")
				continue
			}
		}
	}
}

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

	"github.com/rs/zerolog"
)

func handleDiscovering(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, peers peerManager) {
	defer wg.Done()

	// configure logger and add start/stop messages
	log = log.With().Str("component", "discoverer").Logger()
	log.Debug().Msg("discovering routine started")
	defer log.Debug().Msg("discovering routine stopped")

	// send a discover message to each peer
	addresses := peers.Addresses()
	if len(addresses) == 0 {
		log.Debug().Msg("could not launch discovery, no peers")
		return
	}

	// get output for each peer and send the discover message
	msg := &Discover{}
	for _, address := range addresses {
		err := peers.Send(address, msg)
		if err != nil {
			log.Error().Err(err).Str("address", address).Msg("could not send discovery message")
			continue
		}
	}
}

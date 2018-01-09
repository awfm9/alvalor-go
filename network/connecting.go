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
// GNU Affero General Public License for more detailb.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package network

import "github.com/rs/zerolog/log"

func checkConnecting() {
	// if we already reached the minimum number of peers, we can abort
	peerCount := mgr.PeerCount()
	if peerCount >= minPeers {
		log.Debug().Msg("no free slots for outgoing peers")
		return
	}

	// we use the difference between min and max peers as available pending
	// connection slots; if there are none free, we can abort
	pendingCount := mgr.PendingCount()
	if peerCount+pendingCount >= maxPeers {
		log.Debug().Msg("no free slots for pending connections")
		return
	}

	// try to get an address to connect to
	address, err := mgr.GetAddress()
	if err != nil {
		log.Error().Err(err).Msg("could not get address")
		return
	}
}

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

package message

import (
	"sync"
)

// The GetInv is a message sent by peers who want to download the given
// block inventory from us. If we have it, we send it to them.
// TODO: reply with not available if we don't have it
func (handler *Handler) processGetInv(wg *sync.WaitGroup, address string, getInv *GetInv) {
	defer wg.Done()

	// configure logger
	with := handler.log.With()
	with.Str("component", "message")
	with.Str("message_type", "get_inv")
	with.Str("address", address)
	with.Hex("hash", getInv.Hash[:])
	log := with.Logger()

	// wrap routine in start and stop messages
	log.Debug().Msg("routine started")
	defer log.Debug().Msg("routine stopped")

	// try to get the inventory
	inv, err := handler.inventories.Get(getInv.Hash)
	if err != nil {
		log.Error().Err(err).Msg("could not get inventory")
		return
	}

	// try to send the inventory
	err = handler.net.Send(address, inv)
	if err != nil {
		log.Error().Err(err).Msg("could not send inventory")
		return
	}

	log.Debug().Msg("processed get_inv message")
}

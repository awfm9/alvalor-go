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

	"github.com/alvalor/alvalor-go/types"
)

func (handler *Handler) processGetInv(wg *sync.WaitGroup, address string, getInv *GetInv) {
	defer wg.Done()

	// configure logger
	log := handler.log.With().Str("component", "message").Str("address", address).Logger()
	log.Debug().Msg("message routine started")
	defer log.Debug().Msg("message routine stopped")

	log = log.With().Str("msg_type", "get_inv").Hex("hash", getInv.Hash[:]).Logger()

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

func (handler *Handler) processInventory(wg *sync.WaitGroup, address string, inv *types.Inventory) {
	defer wg.Done()

	// configure logger
	log := handler.log.With().Str("component", "message").Str("address", address).Logger()
	log = log.With().Str("msg_type", "inventory").Hex("hash", inv.Hash[:]).Int("num_hashes", len(inv.Hashes)).Logger()
	log.Debug().Msg("message routine started")
	defer log.Debug().Msg("message routine stopped")

	// cancel any pending download retries for this inventory
	handler.downloads.Cancel(inv.Hash)

	// mark the inventory as received for the respective peer
	handler.peers.Received(address, inv.Hash)

	// store the new inventory in our database
	err := handler.inventories.Add(inv)
	if err != nil {
		log.Error().Err(err).Msg("could not store received inventory")
		return
	}

	// signal the new inventory to the tracker to start pending tx downloads
	err = handler.paths.Signal(inv.Hash)
	if err != nil {
		log.Error().Err(err).Msg("could not signal inventory")
		return
	}

	log.Debug().Msg("processed inventory message")
}

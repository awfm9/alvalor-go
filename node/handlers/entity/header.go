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

package entity

import (
	"sync"

	"github.com/alvalor/alvalor-go/node/state/peers"
	"github.com/alvalor/alvalor-go/types"
)

func (handler *Handler) processHeader(wg *sync.WaitGroup, header *types.Header) {
	defer wg.Done()

	// precompute header hash
	header.Hash = header.GetHash()

	// configure logger
	with := handler.log.With()
	with.Str("component", "entity")
	with.Str("entity_type", "header")
	with.Hex("hash", header.Hash[:])
	log := with.Logger()

	// wrap routine in start and stop message
	log.Debug().Msg("routine started")
	defer log.Debug().Msg("routine stopped")

	// if we already know the header, we ignore it
	ok := handler.headers.Has(header.Hash)
	if ok {
		log.Debug().Msg("header already known")
		return
	}

	// check the validity of the header
	// TODO

	// add the header to the pathfinder
	err := handler.headers.Add(header)
	if err != nil {
		log.Error().Err(err).Msg("could not add header")
		return
	}

	// we let subscribers know that we received a new header
	handler.events.Header(header.Hash)

	// we should propagate it to peers who are unaware of the header
	// TODO: change broadcast to have target addresses and not exclusion
	addresses := handler.peers.Addresses(peers.HasEntity(false, header.Hash))
	err = handler.net.Broadcast(header, addresses...)
	if err != nil {
		log.Error().Err(err).Msg("could not propagate entity")
		return
	}

	// switch the downloader to the new best path
	path, _ := handler.headers.Path()
	err = handler.paths.Follow(path)
	if err != nil {
		log.Error().Err(err).Msg("could not follow changed path")
		return
	}

	log.Debug().Msg("header processed")
}

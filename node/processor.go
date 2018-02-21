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

package node

import (
	"encoding/hex"
	"sync"

	"github.com/rs/zerolog"

	"github.com/alvalor/alvalor-go/types"
)

func handleProcessing(log zerolog.Logger, wg *sync.WaitGroup, entity Entity, pool pool, state stateManager, handlers handlerManager) {
	defer wg.Done()

	var (
		id = entity.ID()
	)

	// configure logger
	log = log.With().Str("component", "processor").Str("id", hex.EncodeToString(id)).Logger()
	log.Info().Msg("processing routine started")
	defer log.Info().Msg("processing routine stopped")

	// process the message according to type
	switch e := entity.(type) {
	case *types.Transaction:
		ok := pool.Known(id)
		if ok {
			log.Debug().Msg("transaction already known")
			return
		}
		err := pool.Add(e)
		if err != nil {
			log.Error().Err(err).Msg("could not add transaction to pool")
			return
		}
		handlers.Propagate(entity)
	}
}

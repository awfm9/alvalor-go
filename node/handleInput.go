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
	"sync"

	"github.com/rs/zerolog"
)

func handleInput(log zerolog.Logger, wg *sync.WaitGroup, handlers handlerManager, subscription <-chan interface{}) {
	defer wg.Done()

	// configure logger
	log = log.With().Str("component", "input").Logger()
	log.Debug().Msg("input routine started")
	defer log.Debug().Msg("input routine stopped")

	for event := range subscription {
		handlers.Event(event)
	}
}

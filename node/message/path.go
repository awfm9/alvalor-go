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

func (handler *Handler) processPath(wg *sync.WaitGroup, address string, path *Path) {
	defer wg.Done()

	// configure logger
	log := handler.log.With().Str("component", "message").Str("address", address).Logger()
	log = log.With().Str("msg_type", "path").Int("num_headers", len(path.Headers)).Logger()
	log.Debug().Msg("message routine started")
	defer log.Debug().Msg("message routine stopped")

	for _, header := range path.Headers {
		handler.entity.Process(wg, header)
	}

	log.Debug().Msg("processed path message")
}

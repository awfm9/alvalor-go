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

func handleInnerSubscriber(log zerolog.Logger, wg *sync.WaitGroup, innerSubscriber chan interface{}, subscribers []subscriber) {
	defer wg.Done()

	log = log.With().Str("component", "innerSubscriber").Logger()
	log.Debug().Msg("handleInnerSubscriber routine started")
	defer log.Debug().Msg("handleInnerSubscriber routine stopped")
Loop:
	for {
		select {
		case msg, ok := <-innerSubscriber:
			if !ok {
				break Loop
			}
			for _, subscriber := range subscribers {
				select {
				case subscriber.buffer <- msg:
				case <-time.After(10 * time.Millisecond):
					log.Debug().Msg("innerSubscriber buffer is stalling")
				}
			}
		}
	}
}

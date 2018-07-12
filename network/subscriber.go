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

func handleSubscriber(log zerolog.Logger, wg *sync.WaitGroup, sub subscriber) {
	defer wg.Done()

	log = log.With().Str("component", "subscriber").Logger()
	log.Debug().Msg("subscriber routine started")
	defer log.Debug().Msg("subscriber routine stopped")
Loop:
	for {
		select {
		case msg, ok := <-sub.buffer:
			if !ok {
				break Loop
			}
			if len(sub.filters) == 0 {
				//Zero filters is same as any filter
				if triggerSubscriber(log, sub.channel, msg, AnyMsgFilter()) {
					continue
				}
			}
			for _, filter := range sub.filters {
				if triggerSubscriber(log, sub.channel, msg, filter) {
					continue
				}
			}
		}
	}
}

func triggerSubscriber(log zerolog.Logger, subscriber chan<- interface{}, msg interface{}, filter func(interface{}) bool) bool {
	if filter(msg) {
		select {
		case subscriber <- msg:
		case <-time.After(10 * time.Millisecond):
			log.Debug().Msg("subscriber is stalling")
		}
		return true
	}
	return false
}

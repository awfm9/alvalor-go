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
	"github.com/rs/zerolog"
	"time"
)

func handleSubscriber(log zerolog.Logger, subscriber chan interface{}, subscribers map[string][]chan<- interface{}) {
	log = log.With().Str("component", "subscriber").Logger()
	log.Debug().Msg("subscriber routine started")
	defer log.Debug().Msg("subscriber routine stopped")
Loop:
	for {
		select {
		case message, ok := <-subscriber:
			if !ok {
				break Loop
			}
			switch msg := message.(type) {
			case *Disconnected:
				triggerSubscribers(log, subscribers, msg, msg.Address)
			case *Connected:
				triggerSubscribers(log, subscribers, msg, msg.Address)
			case *Received:
				triggerSubscribers(log, subscribers, msg, msg.Address)
			default:
			}
		}
	}
}

func triggerSubscribers(log zerolog.Logger, subscribers map[string][]chan<- interface{}, msg interface{}, addr string) {
	activeSubscribers := append(subscribers[addr], subscribers[""]...)
	duplicateLookup := make(map[chan<- interface{}]struct{})
	for _, activeSubscriber := range activeSubscribers {
		_, ok := duplicateLookup[activeSubscriber]
		if !ok {
			select {
			case activeSubscriber <- msg:
			case <-time.After(10 * time.Millisecond):
				log.Debug().Msg("subscriber is stalling")
			}
		}
		duplicateLookup[activeSubscriber] = struct{}{}
	}
}

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

func handleProcessing(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, book addressManager, events eventManager, address string, input <-chan interface{}, output chan<- interface{}) {
	defer wg.Done()

	// configuration parameters
	var (
		interval = cfg.interval
	)

	// configure logger and add start/stop messages
	log = log.With().Str("component", "processor").Str("address", address).Logger()
	log.Debug().Msg("processing routine started")
	defer log.Debug().Msg("processing routine stopped")

	// the timeout is set to the duration of three heartbeats plus a bit
	timeout := time.Duration(3.5 * float64(interval))

	// start with a discover so we get a picture of the network
	output <- &Discover{}

	// keep processing incoming messages & reply adequately
Loop:
	for {
		select {

		// if the cascade arrives (input is closed) we break the loop
		case message, ok := <-input:
			if !ok {
				break Loop
			}

			// if we receive a message, we process it adequately depending on type
			switch msg := message.(type) {
			case *Ping:
				log.Debug().Msg("ping received")
				output <- &Pong{}
			case *Pong:
				log.Debug().Msg("pong received")
			case *Discover:
				log.Debug().Msg("discover received")
				sample := book.Sample(8)
				output <- &Peers{Addresses: sample}
			case *Peers:
				log.Debug().Msg("peer received")
				for _, address := range msg.Addresses {
					book.Add(address)
				}

			// custom messages should go to the subscriber, but we drop it if the subscriber is stalling
			default:
				log.Debug().Msg("custom received")
				err := events.Received(address, message)
				if err != nil {
					log.Error().Err(err).Msg("could not submit received event")
					continue
				}
			}

		// if this case is triggered, we didn't receive a message in a while and we can drop the peer
		case <-time.After(timeout):
			log.Debug().Msg("peer timed out")
			break Loop
		}
	}

	// once we are here, we want to wait for the cascade in case we broke due to timeout
	for range input {
	}

	// then we propagate the cascade to the sender
	close(output)
}

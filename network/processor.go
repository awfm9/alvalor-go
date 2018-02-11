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

func handleProcessing(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, addresses addressManager, peers peerManager, subscriber chan<- interface{}, address string, input <-chan interface{}, output chan<- interface{}) {
	defer wg.Done()

	// configuration parameters
	var (
		interval = cfg.interval
		listen   = cfg.listen
		laddress = cfg.address
	)

	// configure logger and add start/stop messages
	log = log.With().Str("component", "processor").Str("address", address).Logger()
	log.Info().Msg("processing routine started")
	defer log.Info().Msg("processing routine stopped")

	// for each message, handle it as adequate
	timeout := time.NewTimer(time.Duration(3.5 * float64(interval)))
	if listen {
		output <- &Peers{Addresses: []string{laddress}}
	}
	output <- &Discover{}
Loop:
	for {
		select {
		case message, ok := <-input:
			if !ok {
				break Loop
			}
			timeout.Stop()
			timeout = time.NewTimer(time.Duration(3.5 * float64(interval)))
			switch msg := message.(type) {
			case *Ping:
				log.Debug().Msg("ping received")
				output <- &Pong{}
			case *Pong:
				log.Debug().Msg("pong received")
			case *Discover:
				log.Debug().Msg("discover received")
				sample := addresses.Sample(8)
				output <- &Peers{Addresses: sample}
			case *Peers:
				log.Debug().Msg("peer received")
				for _, address := range msg.Addresses {
					addresses.Add(address)
				}
			default:
				log.Debug().Msg("custom received")
				received := &Received{
					Address:   address,
					Timestamp: time.Now(),
					Message:   message,
				}
				select {
				case subscriber <- received:
					// success
				default:
					log.Debug().Msg("subscriber stalling")
					// no subscriber of subscriber stalling
				}
			}
		case <-time.After(interval):
			log.Debug().Msg("sending heartbeat")
			output <- &Ping{}
		case <-timeout.C:
			log.Info().Msg("peer timed out, dropping")
			err := peers.Drop(address)
			if err != nil {
				log.Error().Err(err).Msg("could not drop peer")
			} else {
				subscriber <- Disconnected{Address: address, Timestamp: time.Now()}
			}

		}
	}
	close(output)
}

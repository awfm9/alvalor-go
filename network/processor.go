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

type processorInfos interface {
	AddressSample() ([]string, error)
}

type processorActions interface {
	DropPeer(address string) error
}

type processorEvents interface {
	Found(address string)
}

func handleProcessing(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, infos processorInfos, actions processorActions, events processorEvents, address string, input <-chan interface{}, output chan<- interface{}) {
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
	timeout := time.NewTimer(interval * 3)
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
			if !timeout.Stop() {
				<-timeout.C
			}
			timeout.Reset(interval * 3)
			switch msg := message.(type) {
			case *Ping:
				log.Debug().Msg("ping received")
				output <- &Pong{}
			case *Pong:
				log.Debug().Msg("pong received")
			case *Discover:
				log.Debug().Msg("discover received")
				addresses, err := infos.AddressSample()
				if err != nil {
					log.Error().Err(err).Msg("could not get address sample")
					continue
				}
				output <- &Peers{Addresses: addresses}
			case *Peers:
				log.Debug().Msg("peer received")
				for _, address := range msg.Addresses {
					events.Found(address)
				}
			}
		case <-time.After(interval):
			output <- &Ping{}
		case <-timeout.C:
			actions.DropPeer(address)
		}
	}
	close(output)
}

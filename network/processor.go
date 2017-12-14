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

	"github.com/rs/zerolog"
)

// Processor is all the dependencies for a processing routine.
type Processor interface {
}

func handleProcessing(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, mgr Processor, address string, input <-chan interface{}, output chan<- interface{}, subscriber chan<- interface{}) {
	defer wg.Done()
	log = log.With().Str("component", "processor").Str("address", address).Logger()
	log.Info().Msg("processing routine started")
	defer log.Info().Msg("processing routine stopped")
	for message := range input {
		switch msg := message.(type) {
		case *Ping:
			log.Debug().Msg("ping received")
			output <- &Pong{}
		case *Pong:
			log.Debug().Msg("pong received")
		default:
			select {
			case subscriber <- msg:
				log.Debug().Msg("forwarded to subscriber")
			default:
				log.Error().Msg("subscriber timed out")
			}
		}
	}
}

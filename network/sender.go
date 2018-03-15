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
	"io"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func handleSending(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, rep reputationManager, events eventManager, address string, output <-chan interface{}, w io.Writer) {
	defer wg.Done()

	// extract configuration parameters
	var (
		codec    = cfg.codec
		interval = cfg.interval
	)

	// configure logger and add stop/start messages
	log = log.With().Str("component", "sender").Str("address", address).Logger()
	log.Debug().Msg("sending routine started")
	defer log.Debug().Msg("sending routine stopped")

	// we keep reading messages from the output channel and writing them to the network connection
	var msg interface{}
	var ok bool
Loop:
	for {

		// if we don't have a message for a while, we send a heartbeat ping
		select {
		case msg, ok = <-output:
			if !ok {
				break Loop
			}
		case <-time.After(interval):
			msg = &Ping{}
		}

		// send the message, break the loop on closed connection, register other failures
		err := codec.Encode(w, msg)
		if errors.Cause(err) == io.EOF || isClosedErr(err) {
			log.Debug().Msg("network connection closed")
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("could not write message")
			rep.Failure(address)
			continue
		}
	}

	// drain the channel in case we broke on closed connection & wait until cascade arrives
	for range output {
	}

	log.Info().Msg("connection dropped")

	err := events.Disconnected(address)
	if err != nil {
		log.Error().Err(err).Msg("could not submit disconnected event")
	}
}

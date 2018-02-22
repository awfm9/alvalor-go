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

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func handleReceiving(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, rep reputationManager, peers peerManager, address string, r io.Reader, input chan<- interface{}) {
	defer wg.Done()

	// extract configuration as needed
	var (
		codec = cfg.codec
	)

	// configure logger and add start/stop messages
	log = log.With().Str("component", "receiver").Str("address", address).Logger()
	log.Debug().Msg("receiving routine started")
	defer log.Debug().Msg("receiving routine stopped")

	// read all messages from connection and forward on input channel; break if connection closed, notify other errors
	for {
		msg, err := codec.Decode(r)
		if errors.Cause(err) == io.EOF || isClosedErr(err) {
			log.Debug().Msg("network connection closed")
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("could not read message")
			rep.Failure(address)
			continue
		}
		input <- msg
	}

	// at this point, we should drop the peer, so that we don't risk sends on closed channels
	peers.Drop(address)

	// once we had a closed network connection, we get here; cascade the shutdown to the processor
	close(input)
}

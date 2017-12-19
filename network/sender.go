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
	"net"
	"sync"

	"github.com/pierrec/lz4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Sender injects the dependencies needed for the sending routine.
type Sender interface {
	DropPeer(address string) error
}

func handleSending(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, mgr Sender, book Book, conn net.Conn, output <-chan interface{}) {
	defer wg.Done()

	// extract configuration parameters
	var (
		address = conn.RemoteAddr().String()
		codec   = cfg.codec
	)

	// configure logger and add stop/start messages
	log = log.With().Str("component", "sender").Str("address", address).Logger()
	log.Info().Msg("sending routine started")
	defer log.Info().Msg("sending routine stopped")

	// read messages from output channel and write to connection
	writer := lz4.NewWriter(conn)
	for msg := range output {
		err := codec.Encode(writer, msg)
		if errors.Cause(err) == io.EOF {
			log.Info().Msg("network connection closed")
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("could not write message")
			book.Error(address)
			mgr.DropPeer(address)
			continue
		}
	}
	for _ = range output {
		// draining the channel
	}
}

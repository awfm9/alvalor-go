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

// Receiver injects the dependencies needed for the receiving routine.
type Receiver interface {
	DropPeer(address string) error
}

func handleReceiving(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, mgr Receiver, book Book, conn net.Conn, input chan<- interface{}) {
	defer wg.Done()

	// extract configuration as needed
	var (
		address = conn.RemoteAddr().String()
		codec   = cfg.codec
	)

	// configure logger and add start/stop messages
	log = log.With().Str("component", "receiver").Str("address", address).Logger()
	log.Info().Msg("receiving routine started")
	defer log.Info().Msg("receiving routine closed")

	// read all messages from connetion and forward on channel
	reader := lz4.NewReader(conn)
	for {
		msg, err := codec.Decode(reader)
		if errors.Cause(err) == io.EOF {
			log.Info().Msg("network connection closed")
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("reading message failed")
			book.Error(address)
			mgr.DropPeer(address)
			continue
		}
		input <- msg
	}
	close(input)
}

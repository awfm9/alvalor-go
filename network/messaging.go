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
	"github.com/rs/zerolog"
)

func handleSending(log zerolog.Logger, wg *sync.WaitGroup, output <-chan interface{}, codec Codec, conn net.Conn) {
	defer wg.Done()
	address := conn.RemoteAddr().String()
	log = log.With().Str("component", "sender").Str("address", address).Logger()
	log.Info().Msg("message sending routine started")
	defer log.Info().Msg("message sending routine stopped")
	writer := lz4.NewWriter(conn)
	for msg := range output {
		err := codec.Encode(writer, msg)
		if err != nil {
			log.Error().Err(err).Msg("could not write message")
			continue
		}
	}
}

func handleReceiving(log zerolog.Logger, wg *sync.WaitGroup, codec Codec, conn net.Conn, input chan<- interface{}) {
	defer wg.Done()
	address := conn.RemoteAddr().String()
	log = log.With().Str("component", "receiver").Str("address", address).Logger()
	log.Info().Msg("message receiving routine started")
	defer log.Info().Msg("message receiving routine closed")
	reader := lz4.NewReader(conn)
	for {
		msg, err := codec.Decode(reader)
		if err != nil && err == io.EOF {
			log.Info().Msg("network connection closed")
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("reading message failed")
			continue
		}
		input <- msg
	}
}

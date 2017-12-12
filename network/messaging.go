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

	"github.com/pierrec/lz4"
	"github.com/rs/zerolog"
)

func handleSending(log zerolog.Logger, output <-chan interface{}, codec Codec, conn net.Conn) {
	address := conn.RemoteAddr().String()
	writer := lz4.NewWriter(conn)
	for msg := range output {
		err := codec.Encode(writer, msg)
		if err != nil {
			log.Error().Str("address", address).Err(err).Msg("could not write message")
			continue
		}
	}
}

func handleReceiving(log zerolog.Logger, codec Codec, conn net.Conn, input chan<- interface{}) {
	address := conn.RemoteAddr().String()
	reader := lz4.NewReader(conn)
	for {
		msg, err := codec.Decode(reader)
		if err != nil && err == io.EOF {
			log.Info().Str("address", address).Msg("network connection closed")
			break
		}
		if err != nil {
			log.Error().Str("address", address).Err(err).Msg("reading message failed")
			continue
		}
		input <- msg
	}
}

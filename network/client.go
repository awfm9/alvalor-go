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
	"net"

	"github.com/rs/zerolog"
)

// dial will launch a dialer that will keep dialing addresses and outputting
// connections.
func dial(log zerolog.Logger, addresses <-chan string, connections chan<- net.Conn) {
	for address := range addresses {
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			log.Error().Err(err).Str("address", address).Msg("could not resolve address")
			continue
		}
		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			log.Error().Err(err).Str("address", address).Msg("could not dial address")
			continue
		}
		connections <- conn
	}
}

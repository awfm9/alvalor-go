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
	"sync"
	"time"

	"github.com/rs/zerolog"
)

func handleDialing(log zerolog.Logger, wg *sync.WaitGroup, addresses <-chan string, connections chan<- net.Conn) {
	defer wg.Done()
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

func handleListening(log zerolog.Logger, wg *sync.WaitGroup, address string, stop <-chan struct{}, connections chan<- net.Conn) {
	defer wg.Done()
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Error().Err(err).Str("address", address).Msg("could not resolve listen address")
		return
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Error().Err(err).Str("address", address).Msg("could not listen on address")
		return
	}
Loop:
	for {
		select {
		case <-stop:
			break Loop
		default:
		}
		ln.SetDeadline(time.Now().Add(100 * time.Millisecond))
		var conn net.Conn
		conn, err = ln.Accept()
		if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
			continue
		}
		if err != nil {
			log.Error().Err(err).Msg("could not accept connection")
			break
		}
		connections <- conn
	}
	err = ln.Close()
	if err != nil {
		log.Error().Err(err).Msg("could not close listener")
		return
	}
}

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
	"bytes"
	"net"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

func handleListening(log zerolog.Logger, wg *sync.WaitGroup, listen string, network []byte, nonce []byte, stop <-chan struct{}, connections chan<- net.Conn) {
	defer wg.Done()
	log = log.With().Str("component", "listener").Str("listen", listen).Logger()
	log.Info().Msg("connection listening routine started")
	defer log.Info().Msg("connection listening routine stopped")
	addr, err := net.ResolveTCPAddr("tcp", listen)
	if err != nil {
		log.Error().Err(err).Msg("could not resolve listen address")
		return
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Error().Err(err).Msg("could not listen on address")
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
		syn := append(network, nonce...)
		ack := make([]byte, len(syn))
		address := conn.RemoteAddr().String()
		_, err = conn.Write(syn)
		if err != nil {
			log.Error().Str("address", address).Err(err).Msg("could not write syn packet")
			conn.Close()
			continue
		}
		_, err = conn.Read(ack)
		if err != nil {
			log.Error().Str("address", address).Err(err).Msg("could not read ack packet")
			conn.Close()
			continue
		}
		networkIn := syn[:len(network)]
		if !bytes.Equal(networkIn, network) {
			log.Error().Str("address", address).Bytes("network", network).Bytes("network_in", networkIn).Msg("network mismatch")
			conn.Close()
			continue
		}
		nonceIn := syn[len(network):]
		if bytes.Equal(nonceIn, nonce) {
			log.Error().Str("address", address).Bytes("nonce", nonce).Msg("identical nonce")
			conn.Close()
			continue
		}
		connections <- conn
	}
	err = ln.Close()
	if err != nil {
		log.Error().Err(err).Msg("could not close listener")
		return
	}
}

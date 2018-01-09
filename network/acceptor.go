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

	"github.com/rs/zerolog"
)

// Acceptor contains all the dependencies needed to accept a connection.
type Acceptor interface {
	ClaimSlot() error
	ReleaseSlot()
	AddPeer(conn net.Conn, nonce []byte) error
}

type AcceptorEvents interface {
	Error(address string)
	Invalid(address string)
	Success(address string)
}

func handleAccepting(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, mgr Acceptor, events AcceptorEvents, conn net.Conn) {
	defer wg.Done()

	// extract configuration parameters we care about
	var (
		network = cfg.network
		nonce   = cfg.nonce
		address = conn.RemoteAddr().String()
	)

	// set up logging with start/stop messages
	log = log.With().Str("component", "acceptor").Str("address", address).Logger()
	log.Info().Msg("accepting routine started")
	defer log.Info().Msg("accepting routine stopped")

	// first make sure we can claim a connection slot
	err := mgr.ClaimSlot()
	if err != nil {
		log.Error().Err(err).Msg("could not claim connection slot")
		conn.Close()
		return
	}
	defer mgr.ReleaseSlot()

	// execute the handshake on the incoming connection
	ack := append(network, nonce...)
	syn := make([]byte, len(ack))
	_, err = conn.Read(syn)
	if err != nil {
		log.Error().Err(err).Msg("could not read syn packet")
		conn.Close()
		events.Error(address)
		return
	}
	networkIn := syn[:len(network)]
	if !bytes.Equal(networkIn, network) {
		log.Error().Bytes("network", network).Bytes("network_in", networkIn).Msg("network mismatch")
		conn.Close()
		events.Invalid(address)
		return
	}
	nonceIn := syn[len(network):]
	if bytes.Equal(nonceIn, nonce) {
		log.Error().Bytes("nonce", nonce).Msg("identical nonce")
		conn.Close()
		events.Invalid(address)
		return
	}
	_, err = conn.Write(ack)
	if err != nil {
		log.Error().Err(err).Msg("could not write ack packet")
		conn.Close()
		events.Error(address)
		return
	}

	// submit the connection for a new peer creation
	err = mgr.AddPeer(conn, nonceIn)
	if err != nil {
		log.Error().Err(err).Msg("could not add peer")
		conn.Close()
		return
	}

	events.Success(address)
}

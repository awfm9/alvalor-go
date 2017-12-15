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

// Connector are the dependencies connecting routines need.
type Connector interface {
	ClaimSlot() error
	ReleaseSlot()
	AddPeer(net.Conn) error
}

func handleConnecting(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, mgr Connector, address string) {
	defer wg.Done()

	// extract the variables from the config we are interested in
	var (
		network = cfg.network
		nonce   = cfg.nonce
	)

	// configure the component logger and set start/stop messages
	log = log.With().Str("component", "connector").Str("address", address).Logger()
	log.Info().Msg("connecting routine started")
	defer log.Info().Msg("connecting routine stopped")

	// claim a free connection slot and set the release
	err := mgr.ClaimSlot()
	if err != nil {
		log.Error().Err(err).Msg("could not claim slot")
		return
	}
	defer mgr.ReleaseSlot()

	// resolve the address and dial the connection
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Error().Err(err).Msg("could not resolve address")
		return
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Debug().Err(err).Msg("could not dial address")
		return
	}

	// execute the network handshake
	syn := append(network, nonce...)
	ack := make([]byte, len(syn))
	_, err = conn.Write(syn)
	if err != nil {
		log.Error().Err(err).Msg("could not write syn packet")
		conn.Close()
		return
	}
	_, err = conn.Read(ack)
	if err != nil {
		log.Error().Err(err).Msg("could not read ack packet")
		conn.Close()
		return
	}
	networkIn := ack[:len(network)]
	if !bytes.Equal(networkIn, network) {
		log.Error().Bytes("network", network).Bytes("network_in", networkIn).Msg("network mismatch")
		conn.Close()
		return
	}
	nonceIn := ack[len(network):]
	if bytes.Equal(nonceIn, nonce) {
		log.Error().Bytes("nonce", nonce).Msg("identical nonce")
		conn.Close()
		return
	}

	// create the peer for the valid connection
	err = mgr.AddPeer(conn)
	if err != nil {
		log.Error().Err(err).Msg("could not add peer")
		conn.Close()
		return
	}
}

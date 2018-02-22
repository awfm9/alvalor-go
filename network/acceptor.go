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
	"encoding/hex"
	"net"
	"sync"

	"github.com/rs/zerolog"
)

func handleAccepting(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, pending pendingManager, peers peerManager, rep reputationManager, book addressManager, events eventManager, conn net.Conn) {

	// synchronization, configuration & logging
	defer wg.Done()
	var (
		network = cfg.network
		nonce   = cfg.nonce
		address = conn.RemoteAddr().String()
	)
	log = log.With().Str("component", "acceptor").Str("address", address).Logger()
	log.Info().Msg("accepting routine started")
	defer log.Info().Msg("accepting routine stopped")

	// first make sure we can claim a connection slot
	err := pending.Claim(address)
	if err != nil {
		log.Error().Err(err).Msg("could not claim connection slot")
		conn.Close()
		return
	}
	defer pending.Release(address)

	// execute the handshake on the incoming connection
	ack := append(network, nonce...)
	syn := make([]byte, len(ack))
	_, err = conn.Read(syn)
	if err != nil {
		log.Error().Err(err).Msg("could not read syn packet")
		conn.Close()
		rep.Failure(address)
		return
	}
	networkIn := syn[:len(network)]
	if !bytes.Equal(networkIn, network) {
		log.Error().Bytes("network", network).Bytes("network_in", networkIn).Msg("network mismatch")
		conn.Close()
		book.Block(address)
		return
	}
	nonceIn := syn[len(network):]
	if bytes.Equal(nonceIn, nonce) {
		log.Error().Str("nonce", hex.EncodeToString(nonce)).Msg("identical nonce")
		conn.Close()
		book.Block(address)
		return
	}
	_, err = conn.Write(ack)
	if err != nil {
		log.Error().Err(err).Msg("could not write ack packet")
		conn.Close()
		rep.Failure(address)
		return
	}

	// submit the connection for a new peer creation
	err = peers.Add(conn, nonceIn)
	if err != nil {
		log.Error().Err(err).Msg("could not add peer")
		conn.Close()
		return
	}

	err = events.Connected(address)
	if err != nil {
		log.Debug().Err(err).Msg("could not submit connected event")
	}
	rep.Success(address)
}

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
	"sync"

	"github.com/rs/zerolog"
)

func handleConnecting(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, pending pendingManager, peers peerManager, rep reputationManager, book addressManager, dialer dialWrapper, events eventManager, address string) {
	defer wg.Done()

	// extract the variables from the config we are interested in
	var (
		network = cfg.network
		nonce   = cfg.nonce
	)

	// configure the component logger and set start/stop messages
	log = log.With().Str("component", "connector").Str("address", address).Logger()
	log.Debug().Msg("connecting routine started")
	defer log.Debug().Msg("connecting routine stopped")

	// claim a free connection slot and set the release
	err := pending.Claim(address)
	if err != nil {
		log.Error().Err(err).Msg("could not claim slot")
		return
	}
	defer pending.Release(address)

	// resolve the address and dial the connection
	conn, err := dialer.Dial(address)
	if err != nil {
		log.Debug().Err(err).Msg("could not dial address")
		rep.Failure(address)
		return
	}

	// execute the network handshake
	syn := append(network, nonce...)
	ack := make([]byte, len(syn))
	_, err = conn.Write(syn)
	if err != nil {
		log.Error().Err(err).Msg("could not write syn packet")
		conn.Close()
		rep.Failure(address)
		return
	}
	_, err = conn.Read(ack)
	if err != nil {
		log.Error().Err(err).Msg("could not read ack packet")
		conn.Close()
		rep.Failure(address)
		return
	}
	networkIn := ack[:len(network)]
	if !bytes.Equal(networkIn, network) {
		log.Error().Bytes("network", network).Bytes("network_in", networkIn).Msg("network mismatch")
		conn.Close()
		book.Block(address)
		return
	}
	nonceIn := ack[len(network):]
	if bytes.Equal(nonceIn, nonce) {
		log.Error().Hex("nonce", nonce).Msg("identical nonce")
		conn.Close()
		book.Block(address)
		return
	}
	if peers.Known(nonceIn) {
		log.Error().Hex("nonce", nonce).Msg("nonce already known")
		conn.Close()
		book.Block(address)
		return
	}

	// create the peer for the valid connection
	err = peers.Add(conn, nonceIn)
	if err != nil {
		log.Error().Err(err).Msg("could not add peer")
		conn.Close()
		return
	}

	log.Info().Msg("outgoing connection established")

	rep.Success(address)

	err = events.Connected(address)
	if err != nil {
		log.Error().Err(err).Msg("could not submit connected event")
	}
}

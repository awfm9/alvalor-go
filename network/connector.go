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
	log = log.With().Str("component", "connector").Str("address", address).Logger()
	log.Info().Msg("connecting routine started")
	defer log.Info().Msg("connecting routine stopped")
	err := mgr.ClaimSlot()
	if err != nil {
		log.Error().Err(err).Msg("could not claim slot")
		return
	}
	defer mgr.ReleaseSlot()
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Error().Err(err).Msg("could not resolve address")
		return
	}
	conn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Error().Err(err).Msg("could not dial address")
		return
	}
	ack := append(cfg.network, cfg.nonce...)
	syn := make([]byte, len(ack))
	_, err = conn.Read(syn)
	if err != nil {
		log.Error().Err(err).Msg("could not read syn packet")
		conn.Close()
		return
	}
	networkIn := syn[:len(cfg.network)]
	if !bytes.Equal(networkIn, cfg.network) {
		log.Error().Bytes("network", cfg.network).Bytes("network_in", networkIn).Msg("network mismatch")
		conn.Close()
		return
	}
	nonceIn := syn[len(cfg.network):]
	if bytes.Equal(nonceIn, cfg.nonce) {
		log.Error().Bytes("nonce", cfg.nonce).Msg("identical nonce")
		conn.Close()
		return
	}
	_, err = conn.Write(ack)
	if err != nil {
		log.Error().Err(err).Msg("could not write ack packet")
		conn.Close()
		return
	}
	err = mgr.AddPeer(conn)
	if err != nil {
		log.Error().Err(err).Msg("could not add peer")
		conn.Close()
		return
	}
}

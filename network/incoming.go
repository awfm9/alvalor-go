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

	"go.uber.org/zap"
)

type Incoming struct {
	address          string
	network          []byte
	nonce            []byte
	log              *zap.Logger
	peerFactory      *PeerFactory
	acceptConnection func([]byte) bool
	onConnecting     func()
	onConnected      func(peer)
	onError          func(conn net.Conn)
}

// listen will start a listener on the configured network address and hand over incoming network
// connections to the welcome handshake function.
func (incoming *Incoming) listen() {
	_, _, err := net.SplitHostPort(incoming.address)
	if err != nil {
		incoming.log.Error("invalid listen address", zap.Error(err))
		return
	}
	ln, err := net.Listen("tcp", incoming.address)
	if err != nil {
		incoming.log.Error("could not create listener", zap.String("address", incoming.address), zap.Error(err))
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			incoming.log.Error("could not accept connection", zap.Error(err))
			break
		}
		if !incoming.acceptConnection([]byte{}) {
			incoming.log.Debug("too many peers", zap.String("address", conn.RemoteAddr().String()))
			conn.Close()
			return
		}
		go incoming.welcome(conn)
	}
}

// welcome starts an incoming handshake by waiting for the peer's node nonce and network ID and
// comparing it against what we are expecting, then sending our response.
func (incoming *Incoming) welcome(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	incoming.log.Info("adding incoming peer", zap.String("address", addr))
	incoming.onConnecting()
	ack := append(incoming.network, incoming.nonce...)
	syn := make([]byte, len(ack))
	_, err := conn.Read(syn)
	if err != nil {
		incoming.onError(conn)
		return
	}
	code := syn[:len(incoming.network)]
	nonce := syn[len(incoming.network):]
	if !bytes.Equal(code, incoming.network) || bytes.Equal(nonce, incoming.nonce) || !incoming.acceptConnection(nonce) {
		incoming.log.Warn("dropping invalid incoming connection", zap.String("address", addr))
		incoming.onError(conn)
		return
	}
	_, err = conn.Write(ack)
	if err != nil {
		incoming.onError(conn)
		return
	}

	incoming.log.Info("finalizing handshake", zap.String("address", addr))
	p := incoming.peerFactory.create(conn, nonce)
	incoming.onConnected(p)
}

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
	"time"

	"github.com/pierrec/lz4"
	"go.uber.org/zap"
)

type Incoming struct {
	address   string
	network   []byte
	nonce     []byte
	log       *zap.Logger
	codec     Codec
	heartbeat time.Duration
	timeout   time.Duration
	onConnected func(peer)
	onConnecting func()
	acceptConnection func([]byte) bool
	onError func(conn net.Conn)
}

// listen will start a listener on the configured network address and hand over incoming network
// connections to the welcome handshake function.
func (node *Incoming) listen() {
	_, _, err := net.SplitHostPort(node.address)
	if err != nil {
		node.log.Error("invalid listen address", zap.Error(err))
		return
	}
	ln, err := net.Listen("tcp", node.address)
	if err != nil {
		node.log.Error("could not create listener", zap.String("address", node.address), zap.Error(err))
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			node.log.Error("could not accept connection", zap.Error(err))
			break
		}
		if !node.acceptConnection([]byte{}) {
			node.log.Debug("too many peers", zap.String("address", conn.RemoteAddr().String()))
			conn.Close()
			return
		}
		go node.welcome(conn)
	}
}

// welcome starts an incoming handshake by waiting for the peer's node nonce and network ID and
// comparing it against what we are expecting, then sending our response.
func (node *Incoming) welcome(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	node.log.Info("adding incoming peer", zap.String("address", addr))
	node.onConnecting()
	ack := append(node.network, node.nonce...)
	syn := make([]byte, len(ack))
	_, err := conn.Read(syn)
	if err != nil {
		node.onError(conn)
		return
	}
	code := syn[:len(node.network)]
	nonce := syn[len(node.network):]
	if !bytes.Equal(code, node.network) || bytes.Equal(nonce, node.nonce) || !node.acceptConnection(nonce) {
		node.log.Warn("dropping invalid incoming connection", zap.String("address", addr))
		node.onError(conn)
		return
	}
	_, err = conn.Write(ack)
	if err != nil {
		node.onError(conn)
		return
	}
	node.init(conn, nonce)
}

// init will initialize a new peer and add it to our registry after a successful handshake. It
// launches the required receiving go routine and does the initial sharing of our own peer address.
// Finally, it notifies the subscriber that a new connection was established.
func (node *Incoming) init(conn net.Conn, nonce []byte) {
	addr := conn.RemoteAddr().String()
	node.log.Info("finalizing handshake", zap.String("address", addr))
	r := lz4.NewReader(conn)
	w := lz4.NewWriter(conn)
	outgoing := make(chan interface{}, 16)
	incoming := make(chan interface{}, 16)
	p := peer{
		conn:      conn,
		addr:      addr,
		nonce:     nonce,
		r:         r,
		w:         w,
		outgoing:  outgoing,
		incoming:  incoming,
		codec:     node.codec,
		heartbeat: node.heartbeat,
		timeout:   node.timeout,
		hb:        time.NewTimer(node.heartbeat),
	}
	node.onConnected(p)
}

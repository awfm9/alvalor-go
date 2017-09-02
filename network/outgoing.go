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

	"go.uber.org/zap"
)

type Outgoing struct {
	nextConnectionFactory func() string
	onConnecting          func()
	acceptConnection      func([]byte) bool
	onConnected           func(peer)
	onError               func(conn net.Conn)
	log                   *zap.Logger
	network               []byte
	nonce                 []byte
	peerFactory           *PeerFactory
	balance               time.Duration
}

// check will see if we are below minimum or above maximum peer count and add or remove peers as
// needed.
func (node *Outgoing) connect() {
	for {
		next := node.nextConnectionFactory()
		if next != "" {
			node.add(next)
		}
		//TODO: Not sure how count can become > than node.maxPeers

		// if count > node.maxPeers {
		// 	node.remove()
		// }
		time.Sleep(node.balance)
	}
}

// add will try to initialize a new outgoing connection and hand over to the outgoing handshake
// function on success.
func (node *Outgoing) add(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		node.log.Error("could not dial peer", zap.String("address", addr), zap.Error(err))
		return
	}
	go node.handshake(conn)
}

// handshake starts an outgoing handshake by sending the network ID and our node nonce, then
// comparing the reply against our initial message.
func (node *Outgoing) handshake(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	node.log.Info("adding outgoing peer", zap.String("address", addr))
	node.onConnecting()
	syn := append(node.network, node.nonce...)
	_, err := conn.Write(syn)
	if err != nil {
		node.onError(conn)
		return
	}
	ack := make([]byte, len(syn))
	_, err = conn.Read(ack)
	if err != nil {
		node.onError(conn)
		return
	}
	code := ack[:len(node.network)]
	nonce := ack[len(node.network):]
	if !bytes.Equal(code, node.network) || bytes.Equal(nonce, node.nonce) || node.acceptConnection(nonce) {
		node.log.Warn("dropping invalid outgoing connection", zap.String("address", addr))
		node.onError(conn)
		return
	}
	node.log.Info("finalizing handshake", zap.String("address", addr))
	p := node.peerFactory.create(conn, nonce)
	node.onConnected(p)
}

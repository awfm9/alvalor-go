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
	network           []byte
	nonce             []byte
	log               *zap.Logger
	peerFactory       *PeerFactory
	nextAddrToConnect func() string
	onConnecting      func()
	acceptConnection  func([]byte) bool
	onConnected       func(peer)
	onError           func(conn net.Conn)
	balance           time.Duration
}

// check will see if we are below minimum or above maximum peer count and add or remove peers as
// needed.
func (outgoing *Outgoing) connect() {
	for {
		next := outgoing.nextAddrToConnect()
		if next != "" {
			outgoing.add(next)
		}
		//TODO: Not sure how count can become > than node.maxPeers

		// if count > node.maxPeers {
		// 	node.remove()
		// }
		time.Sleep(outgoing.balance)
	}
}

// add will try to initialize a new outgoing connection and hand over to the outgoing handshake
// function on success.
func (outgoing *Outgoing) add(addr string) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		outgoing.log.Error("could not dial peer", zap.String("address", addr), zap.Error(err))
		return
	}
	go outgoing.handshake(conn)
}

// handshake starts an outgoing handshake by sending the network ID and our node nonce, then
// comparing the reply against our initial message.
func (outgoing *Outgoing) handshake(conn net.Conn) {
	addr := conn.RemoteAddr().String()
	outgoing.log.Info("adding outgoing peer", zap.String("address", addr))
	outgoing.onConnecting()
	syn := append(outgoing.network, outgoing.nonce...)
	_, err := conn.Write(syn)
	if err != nil {
		outgoing.onError(conn)
		return
	}
	ack := make([]byte, len(syn))
	_, err = conn.Read(ack)
	if err != nil {
		outgoing.onError(conn)
		return
	}
	code := ack[:len(outgoing.network)]
	nonce := ack[len(outgoing.network):]
	if !bytes.Equal(code, outgoing.network) || bytes.Equal(nonce, outgoing.nonce) || outgoing.acceptConnection(nonce) {
		outgoing.log.Warn("dropping invalid outgoing connection", zap.String("address", addr))
		outgoing.onError(conn)
		return
	}
	outgoing.log.Info("finalizing handshake", zap.String("address", addr))
	p := outgoing.peerFactory.create(conn, nonce)
	outgoing.onConnected(p)
}

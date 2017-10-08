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
	"sync/atomic"
	"time"

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

// Server represents the network component listening for incoming connections
// and performing the initial handshake to make sure we are dealing with a valid
// peer of our configured Alvalor network.
type Server struct {
	log       *zap.Logger
	wg        *sync.WaitGroup
	addresses <-chan string
	events    chan<- interface{}
	running   uint32
	address   string
	network   []byte
	nonce     []byte
}

// NewServer will create a new server to listen for incoming peers and handling
// the handshake up to having a valid network connection for the given Alvalor
// network.
func NewServer(log *zap.Logger, wg *sync.WaitGroup, addresses <-chan string, events chan<- interface{}, options ...func(*Server)) *Server {
	svr := &Server{
		log:       log,
		wg:        wg,
		addresses: addresses,
		events:    events,
		running:   1,
		address:   "",
		network:   []byte{0, 0, 0, 0},
		nonce:     uuid.UUID{}.Bytes(),
	}
	for _, option := range options {
		option(svr)
	}
	wg.Add(1)
	go svr.listen()
	return svr
}

// SetAddress allows us to define the local address we want to listen on with
// the svr.
func SetAddress(address string) func(*Server) {
	return func(svr *Server) {
		svr.address = address
	}
}

// SetServerNetwork allows us to set the network to use during the initial connection
// handshake.
func SetServerNetwork(network []byte) func(*Server) {
	return func(svr *Server) {
		svr.network = network
	}
}

// SetServerNonce allows us to set our node nonce to make sure we never connect to
// ourselves.
func SetServerNonce(nonce []byte) func(*Server) {
	return func(svr *Server) {
		svr.nonce = nonce
	}
}

// Listen will start a listener on the configured network address and do the
// welcome handshake, forwarding valid peer connections.
func (svr *Server) listen() {

	// we parse / try to resolve the TCP address to make sure it's valid
	addr, err := net.ResolveTCPAddr("tcp", svr.address)
	if err != nil {
		svr.log.Error("invalid listen address", zap.String("svr.address", svr.address), zap.Error(err))
		return
	}

	// we use a TCP listener here so that we can set deadlines, which avoid having
	// to block on calls to accept and makes it possible to shutdown cleanly
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		svr.log.Error("could not create listener", zap.String("svr.address", svr.address), zap.Error(err))
		return
	}

Loop:
	for atomic.LoadUint32(&svr.running) > 0 {

		// each second, we check if we have a new connection; if we don't, we have
		// a timeout error and we can just go into a new iteration of the loop,
		// which will check if we still want to be running
		ln.SetDeadline(time.Now().Add(100 * time.Millisecond))
		conn, err := ln.Accept()
		if netErr, ok := err.(*net.OpError); ok && netErr.Timeout() {
			continue
		}
		if err != nil {
			svr.log.Error("could not accept connection", zap.Error(err))
			continue
		}

		// at this point we have a valid incoming TCP connection, and we want to
		// make sure there is still an open slot for peers; we don't really charge
		// about the address we take from the channel, as that is only used for
		// the outgoing attempts
		address := conn.RemoteAddr().String()
		select {
		case <-svr.addresses:
		default:
			svr.log.Info("no available connection slots", zap.String("address", address))
			conn.Close()
			continue Loop
		}

		// now that we have taken an available peer slot, we can execute the
		// handshake and make sure we are on the same network and nonces are valid
		ack := append(svr.network, svr.nonce...)
		syn := make([]byte, len(ack))
		_, err = conn.Read(syn)
		if err != nil {
			svr.log.Error("could not read syn packet", zap.Error(err))
			conn.Close()
			svr.events <- Failure{Address: address}
			continue
		}
		network := syn[:len(svr.network)]
		if !bytes.Equal(network, svr.network) {
			svr.log.Warn("dropping invalid network peer", zap.String("address", address), zap.ByteString("network", network))
			conn.Close()
			svr.events <- Violation{Address: address}
			continue
		}
		nonce := syn[len(svr.network):]
		if bytes.Equal(nonce, svr.nonce) {
			svr.log.Warn("dropping connection to self", zap.String("address", address))
			conn.Close()
			svr.events <- Violation{Address: address}
			continue
		}
		_, err = conn.Write(ack)
		if err != nil {
			svr.log.Error("could not write ack packet", zap.Error(err))
			conn.Close()
			svr.events <- Failure{Address: address}
			continue
		}

		// with the handshake completed, we now have a new connection event for a
		// valid peer with known nonce and address
		svr.events <- Connection{Address: address, Conn: conn, Nonce: nonce}
	}

	// once we break this loop, we want to let the waitgroup know this component
	// is done shutting down
	svr.wg.Done()
}

// Close will stop the execution of the server component.
func (svr *Server) Close() {
	atomic.StoreUint32(&svr.running, 0)
}

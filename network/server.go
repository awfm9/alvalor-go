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

	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
)

// Server represents the network component listening for incoming connections
// and performing the initial handshake to make sure we are dealing with a valid
// peer of our configured Alvalor network.
type Server struct {
	log         *zap.Logger
	full        func() bool
	connections chan<- net.Conn
	address     string
	network     []byte
	nonce       []byte
}

// NewServer will create a new server to listen for incoming peers and handling
// the handshake up to having a valid network connection for the given Alvalor
// network.
func NewServer(log *zap.Logger, full func() bool, connections chan<- net.Conn, options ...func(*Server)) *Server {
	server := &Server{
		log:         log,
		full:        full,
		connections: connections,
		address:     "",
		network:     []byte{0, 0, 0, 0},
		nonce:       uuid.UUID{}.Bytes(),
	}
	for _, option := range options {
		option(server)
	}
	// NOTE: should we launch this in our constructors or in the main file?
	go server.listen()
	return server
}

// SetAddress allows us to define the local address we want to listen on with
// the server.
func SetAddress(address string) func(*Server) {
	return func(server *Server) {
		server.address = address
	}
}

// SetNetwork allows us to set the network to use during the initial connection
// handshake.
func SetNetwork(network []byte) func(*Server) {
	return func(server *Server) {
		server.network = network
	}
}

// SetNonce allows us to set our node nonce to make sure we never connect to
// ourselves.
func SetNonce(nonce []byte) func(*Server) {
	return func(server *Server) {
		server.nonce = nonce
	}
}

// listen will start a listener on the configured network address and do the
// welcome handshake, forwarding only valid peer connections.
func (server *Server) listen() {
	_, _, err := net.SplitHostPort(server.address)
	if err != nil {
		server.log.Error("invalid listen address, aborting", zap.String("address", server.address), zap.Error(err))
		return
	}
	ln, err := net.Listen("tcp", server.address)
	if err != nil {
		server.log.Error("could not create listener, aborting", zap.String("address", server.address), zap.Error(err))
		return
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			server.log.Error("could not accept connection", zap.Error(err))
			continue
		}
		address := conn.RemoteAddr().String()
		if server.full() {
			server.log.Info("node full, dropping connection", zap.String("address", address))
			conn.Close()
			continue
		}
		ack := append(server.network, server.nonce...)
		syn := make([]byte, len(ack))
		_, err = conn.Read(syn)
		if err != nil {
			server.log.Error("could not read syn packet", zap.Error(err))
			conn.Close()
			continue
		}
		network := syn[:len(server.network)]
		if !bytes.Equal(network, server.network) {
			server.log.Warn("dropping invalid network peer", zap.String("address", address), zap.ByteString("network", network))
			conn.Close()
			continue
		}
		nonce := syn[len(server.network):]
		if bytes.Equal(nonce, server.nonce) {
			server.log.Warn("dropping connection to self", zap.String("address", address))
			conn.Close()
			continue
		}
		_, err = conn.Write(ack)
		if err != nil {
			server.log.Error("could not write ack packet", zap.Error(err))
			conn.Close()
			continue
		}
		select {
		case server.connections <- conn:
			server.log.Info("submitted new incoming connection", zap.String("address", address))
		case <-time.After(time.Second):
			server.log.Error("incoming connection submission timed out", zap.String("address", address))
			conn.Close()
			continue
		}
	}
}

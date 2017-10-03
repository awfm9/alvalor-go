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

	uuid "github.com/satori/go.uuid"

	"go.uber.org/zap"
)

// Client represents our local network client who reaches out to create new peer
// connections and forwards valid connections.
type Client struct {
	log         *zap.Logger
	addresses   <-chan string
	events      chan<- Event
	connections chan<- net.Conn
	network     []byte
	nonce       []byte
}

// NewClient creates a new client who manages outgoing network connections.
func NewClient(log *zap.Logger, addresses <-chan string, events chan<- Event, connections chan<- net.Conn, options ...func(*Client)) *Client {
	client := &Client{
		log:         log,
		addresses:   addresses,
		events:      events,
		connections: connections,
		network:     []byte{0, 0, 0, 0},
		nonce:       uuid.UUID{}.Bytes(),
	}
	for _, option := range options {
		option(client)
	}
	return client
}

// SetClientNetwork allows us to set the network to use during the initial connection
// handshake.
func SetClientNetwork(network []byte) func(*Client) {
	return func(client *Client) {
		client.network = network
	}
}

// SetClientNonce allows us to set our node nonce to make sure we never connect to
// ourselves.
func SetClientNonce(nonce []byte) func(*Client) {
	return func(client *Client) {
		client.nonce = nonce
	}
}

// add will try to initialize a new outgoing connection and hand over to the outgoing handshake
// function on success.
func (client *Client) dial() {
	for address := range client.addresses {
		_, _, err := net.SplitHostPort(address)
		if err != nil {
			client.log.Error("invalid outgoing address", zap.String("address", address), zap.Error(err))
			client.events <- Event{Address: address, Type: EventInvalid}
			continue
		}
		conn, err := net.Dial("tcp", address)
		if err != nil {
			client.log.Error("could not dial address", zap.String("address", address), zap.Error(err))
			client.events <- Event{Address: address, Type: EventFailed}
			continue
		}
		syn := append(client.network, client.nonce...)
		_, err = conn.Write(syn)
		if err != nil {
			client.log.Error("could not write syn packet", zap.String("address", address), zap.Error(err))
			conn.Close()
			client.events <- Event{Address: address, Type: EventFailed}
			continue
		}
		ack := make([]byte, len(syn))
		_, err = conn.Read(ack)
		if err != nil {
			client.log.Error("could not read ack packet", zap.String("address", address), zap.Error(err))
			conn.Close()
			client.events <- Event{Address: address, Type: EventFailed}
			continue
		}
		network := syn[:len(client.network)]
		if !bytes.Equal(network, client.network) {
			client.log.Warn("dropping invalid network peer", zap.String("address", address), zap.ByteString("network", network))
			conn.Close()
			client.events <- Event{Address: address, Type: EventInvalid}
			continue
		}
		nonce := syn[len(client.network):]
		if bytes.Equal(nonce, client.nonce) {
			client.log.Warn("dropping connection to self", zap.String("address", address))
			conn.Close()
			client.events <- Event{Address: address, Type: EventInvalid}
			continue
		}
		client.connections <- conn
	}
}

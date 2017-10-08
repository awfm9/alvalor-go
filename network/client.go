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

// Client represents our local network client who reaches out to create new peer
// connections and forwards valid connections.
type Client struct {
	log       *zap.Logger
	wg        *sync.WaitGroup
	addresses <-chan string
	events    chan<- interface{}
	running   uint32
	network   []byte
	nonce     []byte
}

// NewClient creates a new client who manages outgoing network connections.
func NewClient(log *zap.Logger, wg *sync.WaitGroup, addresses <-chan string, events chan<- interface{}, options ...func(*Client)) *Client {
	cli := &Client{
		log:       log,
		wg:        wg,
		addresses: addresses,
		events:    events,
		running:   1,
		network:   []byte{0, 0, 0, 0},
		nonce:     uuid.UUID{}.Bytes(),
	}
	for _, option := range options {
		option(cli)
	}
	wg.Add(1)
	go cli.dial()
	return cli
}

// SetClientNetwork allows us to set the network to use during the initial connection
// handshake.
func SetClientNetwork(network []byte) func(*Client) {
	return func(cli *Client) {
		cli.network = network
	}
}

// SetClientNonce allows us to set our node nonce to make sure we never connect to
// ourselves.
func SetClientNonce(nonce []byte) func(*Client) {
	return func(cli *Client) {
		cli.nonce = nonce
	}
}

// dial will try to initialize a new outgoing connection and hand over to the outgoing handshake
// function on success.
func (cli *Client) dial() {

Loop:
	for atomic.LoadUint32(&cli.running) > 0 {

		// make sure we check for shutdown at least once a second
		var address string
		select {
		case address = <-cli.addresses:
		case <-time.After(100 * time.Millisecond):
			continue Loop
		}

		// we will check if the address is valid before dialing
		addr, err := net.ResolveTCPAddr("tcp", address)
		if err != nil {
			cli.log.Error("invalid outgoing address", zap.String("address", address), zap.Error(err))
			cli.events <- Violation{Address: address}
			continue
		}

		// once we have a valid address, we dial with automatic local address
		conn, err := net.DialTCP("tcp", nil, addr)
		if err != nil {
			cli.log.Error("could not dial address", zap.String("address", address), zap.Error(err))
			cli.events <- Failure{Address: address}
			continue
		}

		// at this point we have a valid network connection and do the handshake
		syn := append(cli.network, cli.nonce...)
		_, err = conn.Write(syn)
		if err != nil {
			cli.log.Error("could not write syn packet", zap.String("address", address), zap.Error(err))
			conn.Close()
			cli.events <- Failure{Address: address}
			continue
		}
		ack := make([]byte, len(syn))
		_, err = conn.Read(ack)
		if err != nil {
			cli.log.Error("could not read ack packet", zap.String("address", address), zap.Error(err))
			conn.Close()
			cli.events <- Failure{Address: address}
			continue
		}
		network := syn[:len(cli.network)]
		if !bytes.Equal(network, cli.network) {
			cli.log.Warn("dropping invalid network peer", zap.String("address", address), zap.ByteString("network", network))
			conn.Close()
			cli.events <- Violation{Address: address}
			continue
		}
		nonce := syn[len(cli.network):]
		if bytes.Equal(nonce, cli.nonce) {
			cli.log.Warn("dropping connection to self", zap.String("address", address))
			conn.Close()
			cli.events <- Violation{Address: address}
			continue
		}

		// after the handshake, we have a valid peer with known address & nonce
		cli.events <- Connection{Address: address, Conn: conn, Nonce: nonce}
	}

	// let the waitgroup know we have shut down
	cli.wg.Done()
}

// Close will shut down the dialing to add outgoing peers.
func (cli *Client) Close() {
	atomic.StoreUint32(&cli.running, 0)
}

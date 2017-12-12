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
	"net"
	"sync"

	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
)

// Enumeration of different networks available. A node configured with one
// network will only successfully connect to nodes of the same network. To be
// used for testing & iteration.
var (
	Odin = []byte{79, 68, 73, 78}
	Thor = []byte{84, 72, 79, 82}
	Loki = []byte{76, 79, 75, 73}
)

// Network represents all the entry points for the network module.
type Network struct {
	cfg          *Config
	stopClient   func()
	stopServer   func()
	stopSender   func()
	stopReceiver func()
}

// New will initialize the network component with the given parameters.
func New(log zerolog.Logger, wg *sync.WaitGroup, options ...func(*Config)) *Network {
	cfg := &Config{
		network:  Odin,
		server:   false,
		address:  "0.0.0.0:31337",
		minPeers: 3,
		maxPeers: 10,
	}
	for _, option := range options {
		option(cfg)
	}
	nonce := uuid.NewV4().Bytes()
	input := make(chan net.Conn)
	output := make(chan net.Conn)
	stop := make(chan struct{})
	connections := make(chan net.Conn)
	addresses := make(chan string)
	messages := make(chan interface{})
	sender := NewSender(log)
	receiver := NewReceiver(log, messages)
	go handleIncoming(log, wg, cfg.network, nonce, input, output)
	go handleOutgoing(log, wg, cfg.network, nonce, input, output)
	if cfg.server {
		wg.Add(1)
		go handleListening(log, wg, cfg.address, stop, connections)
	}
	go handleDialing(log, wg, addresses, connections)
	return &Network{
		cfg:          cfg,
		stopServer:   func() { close(stop) },
		stopClient:   func() { close(addresses) },
		stopSender:   sender.stop,
		stopReceiver: receiver.stop,
	}
}

// Stop will shut down all network related activity.
func (net *Network) Stop() {
	net.stopServer()
	net.stopClient()
	net.stopSender()
	net.stopReceiver()
}

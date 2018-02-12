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
	"io"
	"net"
	"sync"
	"time"

	"github.com/pkg/errors"
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

// Network represents a wrapper around the network package to provide the API.
type Network interface {
	Add(address string)
	Broadcast(i interface{}, exclude ...string)
	Send(address string, i interface{}) error
	Subscribe() <-chan interface{}
	Stop()
}

// simpleNetwork represents a simple network wrapper.
type simpleNetwork struct {
	log        zerolog.Logger
	wg         *sync.WaitGroup
	cfg        *Config
	dialer     dialWrapper
	listener   listenWrapper
	addresses  addressManager
	pending    pendingManager
	peers      peerManager
	rep        reputationManager
	subscriber chan interface{}
	stop       chan struct{}
}

// New will initialize the network component.
func New(log zerolog.Logger, codec Codec, options ...func(*Config)) Network {

	// initialize the launcher for all handlers
	net := &simpleNetwork{}

	// add the package information to the top package level logger
	log = log.With().Str("package", "network").Logger()
	net.log = log

	// initialize the package-wide waitgroup
	wg := &sync.WaitGroup{}
	net.wg = wg

	// initialize the default configuration and apply custom options
	cfg := &Config{
		network:    Odin,
		listen:     false,
		address:    "0.0.0.0:31337",
		minPeers:   3,
		maxPeers:   10,
		maxPending: 16,
		nonce:      uuid.NewV4().Bytes(),
		interval:   time.Second * 1,
		codec:      codec,
		bufferSize: 16,
	}
	for _, option := range options {
		option(cfg)
	}
	net.cfg = cfg

	// initialize the address manager that handles outgoing addresses
	addresses := newSimpleAddressManager()
	addresses.Block(cfg.address)
	net.addresses = addresses

	// initialize the slots manager that handles connection slots
	pending := newSimplePendingManager(cfg.maxPending)
	net.pending = pending

	// initialize the peer manager that handles connected peers
	peers := newSimplePeerManager(net, cfg.minPeers, cfg.maxPeers)
	net.peers = peers

	// initialize the reputation manager that handles reputation of peers
	rep := newSimpleReputationManager()
	net.rep = rep

	// create the subscriber channel
	subscriber := make(chan interface{}, int(cfg.maxPeers*cfg.bufferSize))
	net.subscriber = subscriber

	// create the channel that will shut everything down
	stop := make(chan struct{})
	net.stop = stop

	// initialize the initial handlers
	net.Dropper()
	net.Server()
	net.Dialer()

	return net
}

func (net *simpleNetwork) Dropper() {
	go handleDropping(net.log, net.wg, net.cfg, net.peers, net.stop)
}

func (net *simpleNetwork) Server() {
	go handleServing(net.log, net.wg, net.cfg, net.peers, net, net.stop)
}

func (net *simpleNetwork) Dialer() {
	go handleDialing(net.log, net.wg, net.cfg, net.peers, net.pending, net.addresses, net.rep, net, net.stop)
}

func (net *simpleNetwork) Listener() {
	go handleListening(net.log, net.wg, net.cfg, net, net.listener, net.stop)
}

func (net *simpleNetwork) Discoverer() {
	go handleDiscovering(net.log, net.wg, net.cfg, net.peers)
}

func (net *simpleNetwork) Acceptor(conn net.Conn) {
	go handleAccepting(net.log, net.wg, net.cfg, net.pending, net.peers, net.rep, conn)
}

func (net *simpleNetwork) Connector(address string) {
	go handleConnecting(net.log, net.wg, net.cfg, net.pending, net.peers, net.rep, net.dialer, address)
}

func (net *simpleNetwork) Sender(address string, output <-chan interface{}, w io.Writer) {
	go handleSending(net.log, net.wg, net.cfg, net.peers, net.rep, address, output, w)
}

func (net *simpleNetwork) Processor(address string, input <-chan interface{}, output chan<- interface{}) {
	go handleProcessing(net.log, net.wg, net.cfg, net.addresses, net.peers, net.subscriber, address, input, output)
}

func (net *simpleNetwork) Receiver(address string, r io.Reader, input chan<- interface{}) {
	go handleReceiving(net.log, net.wg, net.cfg, net.peers, net.rep, address, r, input)
}

func (net *simpleNetwork) Add(address string) {
	net.addresses.Add(address)
}

func (net *simpleNetwork) Stop() {
	close(net.stop)
	addresses := net.peers.Addresses()
	for _, address := range addresses {
		net.peers.Drop(address)
	}
	net.wg.Wait()
}

// Subscribe returns a channel that will stream all received messages and events.
func (net *simpleNetwork) Subscribe() <-chan interface{} {
	return net.subscriber
}

// Broadcast broadcasts a message to all peers.
func (net *simpleNetwork) Broadcast(i interface{}, exclude ...string) {
	addresses := net.peers.Addresses()
	lookup := make(map[string]struct{})
	for _, address := range exclude {
		lookup[address] = struct{}{}
	}
	for _, address := range addresses {
		_, ok := lookup[address]
		if ok {
			continue
		}
		output, err := net.peers.Output(address)
		if err != nil {
			net.log.Error().Err(err).Str("address", address).Msg("could not broadcast to peer")
			continue
		}
		output <- i
	}
}

// Send sends a message to the peer with the given address.
func (net *simpleNetwork) Send(address string, i interface{}) error {
	output, err := net.peers.Output(address)
	if err != nil {
		return errors.Wrap(err, "could not send to peer")
	}
	output <- i
	return nil
}

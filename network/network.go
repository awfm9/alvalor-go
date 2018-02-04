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
	"sync"
	"time"

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
	Stop()
}

// New will initialize the network component.
func New(log zerolog.Logger, codec Codec, options ...func(*Config)) Network {

	// initialize the launcher for all handlers
	handlers := &simpleHandlerManager{}

	// add the package information to the top package level logger
	log = log.With().Str("package", "network").Logger()
	handlers.log = log

	// initialize the package-wide waitgroup
	wg := &sync.WaitGroup{}
	handlers.wg = wg

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
	handlers.cfg = cfg

	// initialize the address manager that handles outgoing addresses
	addresses := newSimpleAddressManager()
	addresses.Block(cfg.address)
	handlers.addresses = addresses

	// initialize the slots manager that handles connection slots
	slots := newSimpleSlotManager(cfg.maxPending)
	handlers.slots = slots

	// initialize the peer manager that handles connected peers
	peers := newSimplePeerManager(handlers, cfg.minPeers, cfg.maxPeers)
	handlers.peers = peers

	// initialize the reputation manager that handles reputation of peers
	rep := newSimpleReputationManager()
	handlers.rep = rep

	// create the channel that will shut everything down
	stop := make(chan struct{})
	handlers.stop = stop

	// initialize the initial handlers
	handlers.Drop()
	handlers.Serve()
	handlers.Dial()

	return handlers
}

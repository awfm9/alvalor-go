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
	"math/rand"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Enumeration of different networks available. A node configured with one
// network will only successfully connect to nodes of the same network. To be
// used for testing & iteration.
var (
	Odin = []byte{79, 68, 73, 78}
	Thor = []byte{84, 72, 79, 82}
	Loki = []byte{76, 79, 75, 73}
)

// Network represents the manager of all network components.
type Network struct {
	cfg *Config
	wg  *sync.WaitGroup
	reg *Registry
	snd *Sender
	rcv *Receiver
}

// New will initialize the completely wired up networking dependencies.
func New(log zerolog.Logger, options ...func(*Config)) *Network {

	// add the package information to the top package level logger
	log = log.With().Str("package", "network").Logger()

	// initialize the default configuration and apply custom options
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

	// initialize the package-wide waitgroup
	wg := &sync.WaitGroup{}

	// initialize the network component with all state
	net := &Network{
		cfg: cfg,
		wg:  wg,
		reg: NewRegistry(),
		snd: NewSender(log),
		rcv: NewReceiver(log, nil),
	}

	// initialize the dropper who will drop random connections when too many
	go handleDropping(log, wg, time.NewTicker(time.Second).C, cfg.maxPeers, net.PeerCount, net.DropPeer)

	return net
}

// DropPeer will drop a random peer from our connections.
func (net *Network) DropPeer() error {
	addresses := make([]string, 0, len(net.reg.peers))
	for address := range net.reg.peers {
		addresses = append(addresses, address)
	}
	address := addresses[rand.Int()%len(addresses)]
	err := net.snd.removeOutput(address)
	if err != nil {
		return errors.Wrap(err, "could not remove peer output")
	}
	err = net.rcv.removeInput(address)
	if err != nil {
		return errors.Wrap(err, "could not remove peer input")
	}
	delete(net.reg.peers, address)
	return nil
}

// PeerCount returns the number of successfully connected to peers.
func (net *Network) PeerCount() uint {
	return uint(len(net.reg.peers))
}

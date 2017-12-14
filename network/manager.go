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

// Manager represents the manager of all network components.
type Manager struct {
	log        zerolog.Logger
	wg         *sync.WaitGroup
	cfg        *Config
	reg        *Registry
	snd        *Sender
	rcv        *Receiver
	book       Book
	codec      Codec
	subscriber chan<- interface{}
}

// NewManager will initialize the completely wired up networking dependencies.
func NewManager(log zerolog.Logger, options ...func(*Config)) *Manager {

	// add the package information to the top package level logger
	log = log.With().Str("package", "network").Logger()

	// initialize the package-wide waitgroup
	wg := &sync.WaitGroup{}

	// initialize the default configuration and apply custom options
	cfg := &Config{
		network:  Odin,
		server:   false,
		address:  "0.0.0.0:31337",
		minPeers: 3,
		maxPeers: 10,
		nonce:    uuid.NewV4().Bytes(),
		interval: time.Second,
	}
	for _, option := range options {
		option(cfg)
	}

	// TODO: validate the configuration parameters

	// initialize the network component with all state
	mgr := &Manager{
		log:   log,
		wg:    wg,
		cfg:   cfg,
		reg:   NewRegistry(),
		snd:   NewSender(log),
		rcv:   NewReceiver(log),
		book:  &SimpleBook{},
		codec: &SimpleCodec{},
	}

	// create the universal stop channel
	stop := make(chan struct{})

	// initialize the dropper who will drop random connections when there are too
	// many
	wg.Add(1)
	go handleDropping(log, wg, cfg, mgr, stop)

	// initialize the connector who will start dialing connections when we are not
	// connected to enough peers
	wg.Add(1)
	go handleDialing(log, wg, cfg, mgr, stop)

	// initialize the listener who will accept connections when
	wg.Add(1)
	go handleServing(log, wg, cfg, mgr, stop)

	return mgr
}

// DropPeer will drop a random peer from our connections.
func (mgr *Manager) DropPeer() error {
	addresses := make([]string, 0, len(mgr.reg.peers))
	for address := range mgr.reg.peers {
		addresses = append(addresses, address)
	}
	address := addresses[rand.Int()%len(addresses)]
	err := mgr.snd.removeOutput(address)
	if err != nil {
		return errors.Wrap(err, "could not remove peer output")
	}
	err = mgr.rcv.removeInput(address)
	if err != nil {
		return errors.Wrap(err, "could not remove peer input")
	}
	delete(mgr.reg.peers, address)
	return nil
}

// PeerCount returns the number of successfully connected to peers.
func (mgr *Manager) PeerCount() uint {
	return uint(len(mgr.reg.peers))
}

// PendingCount returns the number of pending peer connections.
func (mgr *Manager) PendingCount() uint {
	return mgr.reg.pending
}

// ClaimSlot claims one pending connection slot.
func (mgr *Manager) ClaimSlot() error {
	mgr.reg.pending++
	return nil
}

// ReleaseSlot releases one pending connection slot.
func (mgr *Manager) ReleaseSlot() {
	mgr.reg.pending--
}

// StartConnector will try to launch a new connection attempt.
func (mgr *Manager) StartConnector() error {
	addresses, err := mgr.book.Sample(1, IsActive(false), RandomSort())
	if err != nil {
		return errors.Wrap(err, "could not get address")
	}
	mgr.wg.Add(1)
	go handleConnecting(mgr.log, mgr.wg, mgr.cfg, mgr, addresses[0])
	return nil
}

// StartListener will start a listener on a given port.
func (mgr *Manager) StartListener(stop <-chan struct{}) error {
	mgr.wg.Add(1)
	go handleListening(mgr.log, mgr.wg, mgr.cfg, mgr, stop)
	return nil
}

// StartAcceptor will start accepting an incoming connection.
func (mgr *Manager) StartAcceptor(conn net.Conn) error {
	mgr.wg.Add(1)
	go handleAccepting(mgr.log, mgr.wg, mgr.cfg, mgr, conn)
	return nil
}

// StartProcessor will start processing on a given connection.
func (mgr *Manager) StartProcessor(conn net.Conn) error {
	address := conn.RemoteAddr().String()
	input, err := mgr.rcv.addInput(conn, mgr.codec)
	if err != nil {
		return errors.Wrap(err, "could not add input")
	}
	output, err := mgr.snd.addOutput(conn, mgr.codec)
	if err != nil {
		return errors.Wrap(err, "could not add output")
	}
	mgr.wg.Add(1)
	go handleProcessing(mgr.log, mgr.wg, mgr.cfg, mgr, address, input, output, mgr.subscriber)
	mgr.reg.peers[address] = &Peer{Address: address, Conn: conn}
	return nil
}

// TODO: add sender and receiver to manager, remove their state

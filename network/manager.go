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
	log     zerolog.Logger
	wg      *sync.WaitGroup
	cfg     *Config
	book    Book
	codec   Codec
	stop    chan struct{}
	conns   map[string]net.Conn
	inputs  map[string]chan interface{}
	outputs map[string]chan interface{}
	pending uint
	known   map[string]string
}

// NewManager will initialize the completely wired up networking dependencies.
func NewManager(log zerolog.Logger, options ...func(*Config)) *Manager {

	// add the package information to the top package level logger
	log = log.With().Str("package", "network").Logger()

	// initialize the package-wide waitgroup
	wg := &sync.WaitGroup{}

	// initialize the default configuration and apply custom options
	cfg := &Config{
		network:    Odin,
		listen:     false,
		address:    "0.0.0.0:31337",
		minPeers:   3,
		maxPeers:   10,
		nonce:      uuid.NewV4().Bytes(),
		interval:   time.Second * 1,
		codec:      SimpleCodec{},
		bufferSize: 16,
	}
	for _, option := range options {
		option(cfg)
	}

	// TODO: validate the configuration parameters

	// initialize the network component with all state
	mgr := &Manager{
		log:     log,
		wg:      wg,
		cfg:     cfg,
		book:    NewSimpleBook(),
		stop:    make(chan struct{}),
		conns:   make(map[string]net.Conn),
		inputs:  make(map[string]chan interface{}),
		outputs: make(map[string]chan interface{}),
	}

	// TODO: separate book package and inject so we can add addresses in main
	mgr.book.Add("127.0.0.1:31330")
	mgr.book.Add("127.0.0.1:31331")
	mgr.book.Add("127.0.0.1:31332")
	mgr.book.Add("127.0.0.1:31333")
	mgr.book.Add("127.0.0.1:31334")
	mgr.book.Add("127.0.0.1:31335")
	mgr.book.Add("127.0.0.1:31336")
	mgr.book.Add("127.0.0.1:31337")
	mgr.book.Add("127.0.0.1:31338")
	mgr.book.Add("127.0.0.1:31339")

	// blacklist our own address
	mgr.book.Invalid(cfg.address)

	// initialize the dropper who will drop random connections when there are too
	// many
	wg.Add(1)
	go handleDropping(log, wg, cfg, mgr, mgr.book, mgr.stop)

	// initialize the connector who will start dialing connections when we are not
	// connected to enough peers
	wg.Add(1)
	go handleDialing(log, wg, cfg, mgr, mgr.stop)

	// initialize the listener who will accept connections when
	wg.Add(1)
	go handleServing(log, wg, cfg, mgr, mgr.stop)

	return mgr
}

// Stop will shut down all routines and wait for them to end.
func (mgr *Manager) Stop() {
	close(mgr.stop)
	mgr.wg.Wait()
}

// GetAddresses will return the addresses of all connected peers.
func (mgr *Manager) GetAddresses() []string {
	addresses := make([]string, 0, len(mgr.conns))
	for address := range mgr.conns {
		addresses = append(addresses, address)
	}
	return addresses
}

// DropPeer will drop a random peer from our connections.
func (mgr *Manager) DropPeer(address string) error {
	conn, ok := mgr.conns[address]
	if !ok {
		return errors.New("peer not found")
	}
	output := mgr.outputs[address]
	delete(mgr.outputs, address)
	close(output)
	input := mgr.inputs[address]
	delete(mgr.inputs, address)
	close(input)
	delete(mgr.conns, address)
	err := conn.Close()
	if err != nil {
		return errors.Wrap(err, "could not close connection")
	}
	for key := range mgr.known {
		if key == address {
			delete(mgr.known, address)
			break
		}
	}
	return nil
}

// PeerCount returns the number of successfully connected to peers.
func (mgr *Manager) PeerCount() uint {
	return uint(len(mgr.conns))
}

// PendingCount returns the number of pending peer connections.
func (mgr *Manager) PendingCount() uint {
	return mgr.pending
}

// KnownNonce will let us know if we are already connected to a peer with the
// given nonce.
func (mgr *Manager) KnownNonce(nonce []byte) bool {
	_, ok := mgr.known[string(nonce)]
	return ok
}

// ClaimSlot claims one pending connection slot.
func (mgr *Manager) ClaimSlot() error {
	mgr.pending++
	return nil
}

// ReleaseSlot releases one pending connection slot.
func (mgr *Manager) ReleaseSlot() {
	mgr.pending--
}

// GetAddress returns a random address for connection.
func (mgr *Manager) GetAddress() (string, error) {
	addresses, err := mgr.book.Sample(1, IsActive(false), RandomSort())
	if err != nil {
		return "", errors.Wrap(err, "could not get address")
	}
	return addresses[0], nil
}

// StartConnector will try to launch a new connection attempt.
func (mgr *Manager) StartConnector(address string) {
	mgr.wg.Add(1)
	go handleConnecting(mgr.log, mgr.wg, mgr.cfg, mgr, mgr.book, address)
}

// StartListener will start a listener on a given port.
func (mgr *Manager) StartListener(stop <-chan struct{}) {
	mgr.wg.Add(1)
	go handleListening(mgr.log, mgr.wg, mgr.cfg, mgr, stop)
}

// StartAcceptor will start accepting an incoming connection.
func (mgr *Manager) StartAcceptor(conn net.Conn) {
	mgr.wg.Add(1)
	go handleAccepting(mgr.log, mgr.wg, mgr.cfg, mgr, mgr.book, conn)
}

// AddPeer will launch all necessary processing for a new valid connection.
func (mgr *Manager) AddPeer(conn net.Conn, nonce []byte) error {

	// check that we have nothing for this connection saved yet
	address := conn.RemoteAddr().String()
	_, ok := mgr.conns[address]
	if ok {
		return errors.New("peer already exists")
	}

	// add the nonce to the list of stuff we know
	mgr.known[string(nonce)] = address

	// save the references needed for later interactions
	input := make(chan interface{}, mgr.cfg.bufferSize)
	output := make(chan interface{}, mgr.cfg.bufferSize)
	mgr.conns[address] = conn
	mgr.inputs[address] = input
	mgr.outputs[address] = output

	// launch the message processing routines
	mgr.wg.Add(1)
	go handleSending(mgr.log, mgr.wg, mgr.cfg, mgr, mgr.book, conn, output)
	mgr.wg.Add(1)
	go handleReceiving(mgr.log, mgr.wg, mgr.cfg, mgr, mgr.book, conn, input)
	mgr.wg.Add(1)
	go handleProcessing(mgr.log, mgr.wg, mgr.cfg, mgr, address, input, output)

	return nil
}

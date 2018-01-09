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
	"time"

	"github.com/pierrec/lz4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	log      zerolog.Logger
	wg       *sync.WaitGroup
	cfg      *Config
	book     *Book
	registry Registry
	stop     chan struct{}
	pending  uint
	handlers []func()
	mutex    sync.Mutex
	listener net.Listener
}

// NewManager will initialize the completely wired up networking dependencies.
func NewManager(log zerolog.Logger, codec Codec, options ...func(*Config)) *Manager {

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
		codec:      codec,
		bufferSize: 16,
	}
	for _, option := range options {
		option(cfg)
	}

	// TODO: validate the configuration parameters

	// initialize the network component with all state
	mgr := &Manager{
		log:      log,
		wg:       wg,
		cfg:      cfg,
		book:     NewBook(),
		registry: NewSimpleRegistry(),
		stop:     make(chan struct{}),
	}

	// blacklist our own address
	mgr.book.Invalid(cfg.address)

	// register drop handler
	mgr.handlers = append(mgr.handlers, func() { handleDropping(log, wg, cfg, mgr, mgr.book) })

	// register dialing handler
	mgr.handlers = append(mgr.handlers, func() { handleDialing(log, wg, cfg, mgr) })

	// register serving handler
	mgr.handlers = append(mgr.handlers, func() { handleServing(log, wg, cfg, mgr) })

	// initialize the emitter which will start the handlers regularly
	wg.Add(1)
	go handleEmitting(log, wg, cfg, mgr, mgr.stop)

	return mgr
}

// Stop will shut down all routines and wait for them to end.
func (mgr *Manager) Stop() {
	close(mgr.stop)
	for _, address := range mgr.registry.List() {
		go mgr.DropPeer(address)
	}
	mgr.wg.Wait()
}

// GetAddresses will return the addresses of all connected peers.
func (mgr *Manager) GetAddresses() []string {
	return mgr.registry.List()
}

// DropPeer will drop a random peer from our connections.
func (mgr *Manager) DropPeer(address string) error {
	peer, ok := mgr.registry.Get(address)
	if !ok {
		return errors.New("peer not found")
	}
	err := peer.conn.Close()
	if err != nil {
		return errors.Wrap(err, "could not close connection")
	}
	_ = mgr.registry.Remove(address)
	return nil
}

// PeerCount returns the number of successfully connected to peers.
func (mgr *Manager) PeerCount() uint {
	return mgr.registry.Count()
}

// PendingCount returns the number of pending peer connections.
func (mgr *Manager) PendingCount() uint {
	return mgr.pending
}

// KnownNonce will let us know if we are already connected to a peer with the
// given nonce.
func (mgr *Manager) KnownNonce(nonce []byte) bool {
	for _, address := range mgr.registry.List() {
		peer, ok := mgr.registry.Get(address)
		if !ok {
			continue
		}
		if bytes.Equal(peer.nonce, nonce) {
			return true
		}
	}
	return false
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
	addresses, err := mgr.book.Sample(1, isActive(false), byRandom())
	if err != nil {
		return "", errors.Wrap(err, "could not get address")
	}
	return addresses[0], nil
}

// StartHandlers will start the registered handlers.
func (mgr *Manager) Launch() {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
}

// StartConnector will try to launch a new connection attempt.
func (mgr *Manager) StartConnector(address string) {
	mgr.wg.Add(1)
	go handleConnecting(mgr.log, mgr.wg, mgr.cfg, mgr, mgr.book, address)
}

// StartListener will start a listener on a given port.
func (mgr *Manager) StartListener() {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	if mgr.listener != nil {
		return
	}
	addr, err := net.ResolveTCPAddr("tcp", mgr.cfg.address)
	if err != nil {
		log.Error().Err(err).Msg("could not resolve listen address")
		return
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Error().Err(err).Msg("could not listen on address")
		return
	}

	mgr.handlers = append(mgr.handlers, func() { handleListening(mgr.log, mgr.wg, mgr.cfg, listener, mgr) })
}

// StopListener will stop a listener on the configured port.
func (mgr *Manager) StopListener() {
	mgr.mutex.Lock()
	defer mgr.mutex.Unlock()
	if mgr.listener == nil {
		return
	}
	mgr.handlers = mgr.handlers[:len(mgr.handlers)-1]
	mgr.listener.Close()
}

// StartAcceptor will start accepting an incoming connection.
func (mgr *Manager) StartAcceptor(conn net.Conn) {
	mgr.wg.Add(1)
	go handleIncoming(mgr.log, mgr.wg, mgr.cfg, mgr, mgr.book, conn)
}

// AddPeer will launch all necessary processing for a new valid connection.
func (mgr *Manager) AddPeer(conn net.Conn, nonce []byte) {

	// create the peer and add to registry
	peer := &Peer{
		conn:   conn,
		input:  make(chan interface{}, mgr.cfg.bufferSize),
		output: make(chan interface{}, mgr.cfg.bufferSize),
		nonce:  nonce,
	}
	err := mgr.registry.Add(peer)
	if err != nil {
		mgr.log.Error().Err(err).Msg("could not add peer to registry")
		return
	}

	// initialize the readers and writers
	address := conn.RemoteAddr().String()
	r := lz4.NewReader(conn)
	w := lz4.NewWriter(conn)

	// launch the message processing routines
	mgr.wg.Add(3)
	go handleSending(mgr.log, mgr.wg, mgr.cfg, mgr, mgr.book, address, peer.output, w)
	go handleReceiving(mgr.log, mgr.wg, mgr.cfg, mgr, mgr.book, address, r, peer.input)
	go handleProcessing(mgr.log, mgr.wg, mgr.cfg, mgr, mgr.book, address, peer.input, peer.output)
}

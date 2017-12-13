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
	log   zerolog.Logger
	wg    *sync.WaitGroup
	cfg   *Config
	reg   *Registry
	snd   *Sender
	rcv   *Receiver
	book  Book
	codec Codec
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

	// initialize the network component with all state
	mgr := &Manager{
		log:   log,
		wg:    wg,
		cfg:   cfg,
		reg:   NewRegistry(),
		snd:   NewSender(log),
		rcv:   NewReceiver(log, nil),
		book:  &SimpleBook{},
		codec: &SimpleCodec{},
	}

	// create the universal stop channel
	stop := make(chan struct{})
	messages := make(chan Message)

	// initialize the dropper who will drop random connections when there are too
	// many
	wg.Add(1)
	go handleDropping(log, wg, cfg, mgr, stop)

	// initialize the connector who will start dialing connections when we are not
	// connected to enough peers
	wg.Add(1)
	go handleDialing(log, wg, cfg, mgr, stop)

	// initialize the processor who will process all incoming messages according
	// to the rules of the protocol
	wg.Add(1)
	go handleProcessing(log, wg, cfg, mgr, messages)

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

// DialConn will try to launch a new connection attempt.
func (mgr *Manager) DialConn() error {
	addresses, err := mgr.book.Sample(1, IsActive(false), RandomSort())
	if err != nil {
		return errors.Wrap(err, "could not get address")
	}
	mgr.wg.Add(1)
	go handleConnecting(mgr.log, mgr.wg, mgr.cfg, mgr, addresses[0])
	return nil
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

// AddPeer will add a new successful peer connection.
func (mgr *Manager) AddPeer(conn net.Conn) error {
	address := conn.RemoteAddr().String()
	err := mgr.rcv.addInput(conn, mgr.codec)
	if err != nil {
		return errors.Wrap(err, "could not add input")
	}
	err = mgr.snd.addOutput(conn, mgr.codec)
	if err != nil {
		return errors.Wrap(err, "could not add output")
	}
	mgr.reg.peers[address] = Peer{Address: address}
	return nil
}

// Peer returns the peer with the given address.
func (mgr *Manager) Peer(address string) (Peer, error) {
	peer, ok := mgr.reg.peers[address]
	if !ok {
		return Peer{}, errors.New("peer not found")
	}
	return peer, nil
}

// Protocol returns the protocol of the given version.
func (mgr *Manager) Protocol(version string) (Protocol, error) {
	return VersionOne{}, nil
}

// Send will send a message to the given peer.
func (mgr *Manager) Send(address string, message interface{}) error {
	return mgr.snd.Send(address, message)
}

// State returns the current local state.
func (mgr *Manager) State() (State, error) {
	return State{}, nil
}

// TODO: how to really do the processing of the messages by the protocol;
// the structure could be quite a bit better maybe

// TODO: how to connect the very rich blockchain state to the local small state
// variable as we modeled it now

// TODO: when injecting dependencies, sometimes manager just proxies, how can
// we improve this? would be cool to have composition or filtering such as in
// functional languages :<

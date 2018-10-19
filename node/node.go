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

package node

import (
	"io"
	"sync"

	"github.com/rs/zerolog"

	"github.com/alvalor/alvalor-go/types"
)

// Network defines what we need from the network module.
type Network interface {
	Send(address string, msg interface{}) error
	Broadcast(msg interface{}, exclude ...string) error
}

// Codec is responsible for serializing and deserializing data for disk storage.
type Codec interface {
	Encode(w io.Writer, i interface{}) error
	Decode(r io.Reader) (interface{}, error)
}

type subscriber struct {
	channel chan<- interface{}
	filters []func(interface{}) bool
}

// Node represents the second layer of the network stack, which understands the
// semantics of entities in the blockchain database and manages synchronization
// of the consensus state.
type Node struct {
	log          zerolog.Logger
	wg           *sync.WaitGroup
	codec        Codec
	net          Network
	paths        Paths
	downloads    Downloads
	events       Events
	headers      Headers
	inventories  Inventories
	transactions Transactions
	peers        Peers
	stop         chan struct{}
}

// New creates a new node to manage the Alvalor blockchain.
func New(log zerolog.Logger, codec Codec, net Network, paths Paths, downloads Downloads, events Events, headers Headers, inventories Inventories, transactions Transactions, peers Peers, input <-chan interface{}) *Node {

	// initialize the node
	n := &Node{}

	// configure the logger
	log = log.With().Str("package", "node").Logger()
	n.log = log

	// initialize the wait group
	wg := &sync.WaitGroup{}
	n.wg = wg

	// store references for component dependencies
	n.codec = codec
	n.net = net

	// store references for helper dependencies
	n.paths = paths
	n.downloads = downloads
	n.events = events

	// store references for repository dependencies
	n.inventories = inventories
	n.headers = headers
	n.transactions = transactions

	// store references for state dependencies
	n.peers = peers

	// start handling input messages from the network layer
	n.input(input)

	return n
}

// Subscribe adds a subscriber to the node, sending events that pass the given
// filters to the provided subscription channel. If the subscriber already
// exists, it will change the filters.
func (n *Node) Subscribe(sub chan<- interface{}, filters ...func(interface{}) bool) {
	n.events.Subscribe(sub, filters...)
}

// Unsubscribe removes the subscriber with the given subscription channel from
// the node.
func (n *Node) Unsubscribe(sub chan<- interface{}) {
	n.events.Unsubscribe(sub)
}

// Submit will submit a transaction to the node for processing.
func (n *Node) Submit(tx *types.Transaction) {
	n.entity(tx)
}

// Stop will wait for all pending handlers to finish.
func (n *Node) Stop() {
	n.wg.Wait()
}

func (n *Node) input(input <-chan interface{}) {
	n.wg.Add(1)
	go handleInput(n.log, n.wg, n, input)
}

func (n *Node) event(event interface{}) {
	n.wg.Add(1)
	go handleEvent(n.log, n.wg, n.net, n.headers, n.peers, n, event)
}

func (n *Node) message(address string, message interface{}) {
	n.wg.Add(1)
	go handleMessage(n.log, n.wg, n.net, n.paths, n.downloads, n.headers, n.inventories, n.transactions, n.peers, n, address, message)
}

func (n *Node) entity(entity types.Entity) {
	n.wg.Add(1)
	go handleEntity(n.log, n.wg, n.net, n.paths, n.events, n.headers, n.transactions, n.peers, n, entity)
}

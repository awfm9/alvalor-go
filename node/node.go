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
	"sync"

	"github.com/rs/zerolog"

	"github.com/alvalor/alvalor-go/node/handler"
	"github.com/alvalor/alvalor-go/types"
)

type subscriber struct {
	channel chan<- interface{}
	filters []func(interface{}) bool
}

// Node represents the second layer of the network stack, which understands the
// semantics of entities in the blockchain database and manages synchronization
// of the consensus state.
type Node struct {
	log     zerolog.Logger
	wg      *sync.WaitGroup
	events  Events
	event   func(interface{})
	message func(string, interface{})
	entity  func(types.Entity)
	stop    chan struct{}
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

	// store references for helper dependencies
	n.events = events

	// bind the handlers
	n.entity = handler.Entity(log, wg, net, paths, events, headers, transactions, peers)
	n.message = handler.Message(log, wg, net, paths, downloads, headers, inventories, transactions, peers, n.entity)
	n.event = handler.Event(log, wg, net, headers, peers, n.message)

	// start handling input messages from the network layer
	n.process(input)

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

// process will start processing the input queue.
func (n *Node) process(input <-chan interface{}) {
	n.wg.Add(1)
	defer n.wg.Done()
	for event := range input {
		n.event(event)
	}
}

// Submit will submit a transaction to the node for processing.
func (n *Node) Submit(tx *types.Transaction) {
	n.wg.Add(1)
	n.entity(tx)
}

// Stop will wait for all pending handlers to finish.
func (n *Node) Stop() {
	n.wg.Wait()
}

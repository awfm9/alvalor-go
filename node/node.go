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

	"github.com/alvalor/alvalor-go/node/peer"
	"github.com/alvalor/alvalor-go/node/repo"
	"github.com/alvalor/alvalor-go/trie"
	"github.com/alvalor/alvalor-go/types"
)

// Node defines the exposed API of the Alvalor node package.
type Node interface {
	Submit(tx *types.Transaction)
	Stats()
}

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

type simpleNode struct {
	log          zerolog.Logger
	wg           *sync.WaitGroup
	net          Network
	download     downloader
	inventories  Inventories
	headers      Headers
	transactions Transactions
	track        tracker
	state        State
	events       eventManager
	stream       chan interface{}
	subscribers  chan *subscriber
	stop         chan struct{}
}

// New creates a new node to manage the Alvalor blockchain.
func New(log zerolog.Logger, net Network, headers Headers, inventories Inventories, codec Codec, input <-chan interface{}) Node {

	// initialize the node
	n := &simpleNode{}

	// configure the logger
	log = log.With().Str("package", "node").Logger()
	n.log = log

	// initialize the wait group
	wg := &sync.WaitGroup{}
	n.wg = wg

	// store references for reused dependencies
	n.net = net

	// create the event and ubscriber streams
	n.stream = make(chan interface{}, 128)
	n.subscribers = make(chan *subscriber)

	// initialize the various data stores
	n.inventories = repo.NewInventories()
	n.headers = headers
	n.state = peer.NewState()

	// initialize the event manager to create events
	n.events = newEventManager(n.stream)

	// initialize simple transaction pool
	n.transactions = repo.NewTransactions(codec, trie.NewBin())

	// start streaming generated events to subscribers
	n.Stream()

	// start handling input messages from the network layer
	n.Input(input)

	return n
}

func (n *simpleNode) Subscribe(channel chan<- interface{}, filters ...func(interface{}) bool) {
	sub := &subscriber{filters: filters, channel: channel}
	n.subscribers <- sub
}

func (n *simpleNode) Submit(tx *types.Transaction) {
	n.Entity(tx)
}

func (n *simpleNode) Stats() {
	numActive := n.state.Count(peer.IsActive(true))
	numPool := uint(len(n.transactions.Pending()))
	n.log.Info().Uint("num_active", numActive).Uint("num_pool", numPool).Msg("stats")
}

func (n *simpleNode) Input(input <-chan interface{}) {
	n.wg.Add(1)
	go handleInput(n.log, n.wg, n, input)
}

func (n *simpleNode) Event(event interface{}) {
	n.wg.Add(1)
	go handleEvent(n.log, n.wg, n.net, n.headers, n.state, n, event)
}

func (n *simpleNode) Message(address string, message interface{}) {
	n.wg.Add(1)
	go handleMessage(n.log, n.wg, n.net, n.download, n.state, n.inventories, n.transactions, n.headers, n.track, n, address, message)
}

func (n *simpleNode) Entity(entity Entity) {
	n.wg.Add(1)
	go handleEntity(n.log, n.wg, n.net, n.headers, n.state, n.transactions, n.track, entity, n.events, n)
}

func (n *simpleNode) Stream() {
	n.wg.Add(1)
	go handleStream(n.log, n.wg, n.stream, n.subscribers)
}

func (n *simpleNode) Stop() {
	close(n.subscribers)
	close(n.stream)
	n.wg.Wait()
}

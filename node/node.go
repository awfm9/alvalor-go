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

type simpleNode struct {
	log         zerolog.Logger
	wg          *sync.WaitGroup
	net         Network
	chain       blockchain
	finder      pathfinder
	downloader  downloader
	peers       peerManager
	pool        poolManager
	events      eventManager
	stream      chan interface{}
	subscribers chan *subscriber
	stop        chan struct{}
}

type subscriber struct {
	channel chan<- interface{}
	filters []func(interface{}) bool
}

// New creates a new node to manage the Alvalor blockchain.
func New(log zerolog.Logger, net Network, chain blockchain, codec Codec, input <-chan interface{}) Node {

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
	n.chain = chain

	// create the event stream for subscribers
	n.stream = make(chan interface{}, 128)

	// create the channel to add subscribers for the event stream
	n.subscribers = make(chan *subscriber)

	// initialize path finder to identify best valid path
	n.finder = newSimplePathfinder(n.chain)

	// initialize the event manager to create events
	n.events = newEventManager(n.stream)

	// initialize peer state manager
	n.peers = newPeers()

	// initialize simple transaction pool
	n.pool = newPool(codec, trie.NewBin())

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
	numActive := uint(len(n.peers.Actives()))
	numTxs := n.pool.Count()
	n.log.Info().Uint("num_active", numActive).Uint("num_txs", numTxs).Msg("stats")
}

func (n *simpleNode) Input(input <-chan interface{}) {
	n.wg.Add(1)
	go handleInput(n.log, n.wg, n, input)
}

func (n *simpleNode) Event(event interface{}) {
	n.wg.Add(1)
	go handleEvent(n.log, n.wg, n.net, n.finder, n.peers, n, event)
}

func (n *simpleNode) Message(address string, message interface{}) {
	n.wg.Add(1)
	// TODO: introduce new blockchain interface
	go handleMessage(n.log, n.wg, n.net, n.chain, n.finder, n.downloader, n, address, message)
}

func (n *simpleNode) Entity(entity Entity) {
	n.wg.Add(1)
	// TODO: insert downloader parameter
	go handleEntity(n.log, n.wg, n.net, n.finder, n.peers, n.pool, nil, entity, n.events, n)
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

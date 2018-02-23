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

	"github.com/alvalor/alvalor-go/trie"
	"github.com/alvalor/alvalor-go/types"
)

// Node defines the exposed API of the Alvalor node package.
type Node interface {
	Submit(tx *types.Transaction)
	Stats()
}

type simpleNode struct {
	log   zerolog.Logger
	wg    *sync.WaitGroup
	net   networkManager
	state stateManager
	pool  poolManager
}

// New creates a new node to manage the Alvalor blockchain.
func New(log zerolog.Logger, net networkManager, codec Codec, subscription <-chan interface{}) Node {

	// initialize the node
	n := &simpleNode{}

	// configure the logger
	log = log.With().Str("package", "node").Logger()
	n.log = log

	// initialize the wait group
	wg := &sync.WaitGroup{}
	n.wg = wg

	// store the network manager
	n.net = net

	// initialize peer state manager
	state := newState()
	n.state = state

	// initialize simple transaction pool
	store := trie.New()
	pool := newPool(codec, store)
	n.pool = pool

	// now we want to subscribe to the network layer and process messages
	wg.Add(1)
	go handleInput(log, wg, n, subscription)

	return n
}

func (n *simpleNode) Submit(tx *types.Transaction) {
	n.Entity(tx)
}

func (n *simpleNode) Stats() {
	numActive := uint(len(n.state.Actives()))
	numTxs := n.pool.Count()
	n.log.Info().Uint("num_active", numActive).Uint("num_txs", numTxs).Msg("stats")
}

func (n *simpleNode) Event(event interface{}) {
	n.wg.Add(1)
	go handleEvent(n.log, n.wg, n, n.net, n.state, n.pool, event)
}

func (n *simpleNode) Message(address string, message interface{}) {
	n.wg.Add(1)
	go handleMessage(n.log, n.wg, n, n.net, n.state, n.pool, address, message)
}

func (n *simpleNode) Entity(entity Entity) {
	n.wg.Add(1)
	go handleEntity(n.log, n.wg, n.net, n.state, entity)
}

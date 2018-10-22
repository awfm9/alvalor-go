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
)

// Node represents the second layer of the network stack, which understands the
// semantics of entities in the blockchain database and manages synchronization
// of the consensus state.
type Node struct {
	log   zerolog.Logger
	wg    *sync.WaitGroup
	event func(interface{})
}

// New creates a new node to manage the Alvalor blockchain.
func New(log zerolog.Logger, event func(interface{})) *Node {

	// initialize the node
	n := &Node{
		log:   log.With().Str("package", "node").Logger(),
		wg:    &sync.WaitGroup{},
		event: event,
	}

	return n
}

// Start will start processing an input queue.
func (n *Node) Start(input <-chan interface{}) {
	n.wg.Add(1)
	defer n.wg.Done()
	for event := range input {
		n.event(event)
	}
}

// Stop will wait for all pending handlers to finish.
func (n *Node) Stop() {
	n.wg.Wait()
}

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

	"github.com/alvalor/alvalor-go/types"
)

// Node represents the interface for an outside user of the Alvalor client.
type Node interface {
	Broadcast(tx *types.Transaction) error
}

// New creates a new node to manage the Alvalor blockchain.
func New(log zerolog.Logger, subscription <-chan interface{}) Node {

	// initialize the node
	n := &simpleNode{}

	// configure the logger
	log = log.With().Str("package", "node").Logger()
	n.log = log

	// initialize the wait group
	wg := &sync.WaitGroup{}
	n.wg = wg

	// now we want to subscribe to the network layer and process messages
	wg.Add(1)
	go handleReceiving(log, wg, subscription, nil, nil)

	return n
}

type simpleNode struct {
	log zerolog.Logger
	wg  *sync.WaitGroup
}

func (n *simpleNode) Broadcast(tx *types.Transaction) error {
	return nil
}

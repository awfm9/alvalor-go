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
	"bytes"

	"github.com/alvalor/alvalor-go/types"
)

// PathFinder keeps track of all block headers and the path with most difficulty.
type PathFinder interface {
	Add(block *types.Block)
	Weight(hash types.Hash) uint64
}

type node struct {
	weight uint64
	header *types.Header
	parent *node
}

// SimplePathFinder is a simple pathfinder using caching to efficiently find the longest paths.
type SimplePathFinder struct {
	root  *node
	heads []*node
}

// NewPathFinder creates a new pathfinder using the provided block as source for the path.
func NewPathFinder(source *types.Header) *SimplePathFinder {
	root := &node{weight: 0, header: source, parent: nil}
	path := &SimplePathFinder{
		root:  root,
		heads: []*node{root},
	}
	return path
}

// Add will add a new block header to the pathfinder.
func (pf *SimplePathFinder) Add(header *types.Header) {

	// first try to find the parent in the tree heads
	// if we find it there, we can just replace it
	for i, head := range pf.heads {
		if !bytes.Equal(head.header.Hash[:], header.Parent[:]) {
			continue
		}
		n := &node{weight: head.weight + header.Diff, header: header, parent: head}
		pf.heads[i] = n
		return
	}

	// otherwise, we have to iterate back from all heads
	// TODO
}

// Best will return the head of the longest path.
func (pf *SimplePathFinder) Best() (*types.Header, uint64) {
	best := pf.root
	for _, head := range pf.heads {
		if head.weight <= best.weight {
			continue
		}
		best = head
	}
	return best.header, best.weight
}

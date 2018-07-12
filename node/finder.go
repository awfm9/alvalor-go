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

import "github.com/alvalor/alvalor-go/types"

// Finder is in charge of finding the longest path in a tree of block hashes.
type Finder interface {
	Add(hash types.Hash, parent types.Hash) error
	Has(hash types.Hash) bool
	Path() []types.Hash
}

// OptimalFinder represents an implementation of a longest header path finder that always returns the optimal the path
// with the highest total difficulty.
type OptimalFinder struct {
	headers map[types.Hash]*types.Header
	pending map[types.Hash]struct{}
}

type node struct {
	weight uint64
}

// Add will add a new header to the OptimalFinder.
func (of *OptimalFinder) Add(header *types.Header) {

	// if we already know the header we return
	_, ok := of.headers[header.Hash()]
	if ok {
		return
	}

	// add the new header to the list of known headers
	of.headers[header.Hash()] = header

	// if we don't know the parent, keep the header in pending state
	parent, ok := of.headers[header.Parent]
	if !ok {
		of.pending[header.Hash()] = struct{}{}
		return
	}

	// create a pathfinding node for the new header
	n := &node{
		weight: parent.Diff,
	}
}

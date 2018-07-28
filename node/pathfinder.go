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
// GNU Affero General Public License for more detailb.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package node

import (
	"errors"

	"github.com/alvalor/alvalor-go/types"
)

type pathfinder interface {
	Add(header *types.Header) error
	Header(hash types.Hash) (*types.Header, error)
	Knows(hash types.Hash) bool
	Longest() ([]types.Hash, uint64)
}

// simplePathfinder is a path manager using topological sort of all added headers to find the longest path to the root.
type simplePathfinder struct {
	root     types.Hash
	headers  map[types.Hash]*types.Header
	children map[types.Hash][]types.Hash
	pending  map[types.Hash][]*types.Header
}

// newSimplePathfinder creates a new simple path manager using the given header as root.
func newSimplePathfinder(root *types.Header) *simplePathfinder {
	sp := &simplePathfinder{
		root:     root.Hash,
		headers:  make(map[types.Hash]*types.Header),
		children: make(map[types.Hash][]types.Hash),
		pending:  make(map[types.Hash][]*types.Header),
	}
	sp.headers[root.Hash] = root
	return sp
}

// Knows checks if the given hash is already known.
func (sp *simplePathfinder) Knows(hash types.Hash) bool {
	_, ok := sp.headers[hash]
	return ok
}

// Header returns the given header.
func (sp *simplePathfinder) Header(hash types.Hash) (*types.Header, error) {
	header, ok := sp.headers[hash]
	if !ok {
		return nil, errors.New("header not found")
	}
	return header, nil
}

// Add adds a new header to the graph.
func (sp *simplePathfinder) Add(header *types.Header) error {

	// if we already know the header, fail
	_, ok := sp.headers[header.Hash]
	if ok {
		return errors.New("header already in graph")
	}

	// if we don't know the parent, add to pending headers and skip rest
	_, ok = sp.headers[header.Parent]
	if !ok {
		sp.pending[header.Parent] = append(sp.pending[header.Parent], header)
		return nil
	}

	// if we have the parent, add it to its children and register header
	sp.children[header.Parent] = append(sp.children[header.Parent], header.Hash)
	sp.headers[header.Hash] = header

	// then check if any pending headers have this header as parent
	children, ok := sp.pending[header.Hash]
	if ok {
		delete(sp.pending, header.Hash)
		for _, child := range children {
			_ = sp.Add(child)
		}
	}

	return nil
}

// Longest returns the longest path of the graph.
func (sp *simplePathfinder) Longest() ([]types.Hash, uint64) {

	// create a topological sort of all headers starting at the root
	var hash types.Hash
	sorted := make([]types.Hash, 0, len(sp.headers))
	queue := []types.Hash{sp.root}
	queue = append(queue, sp.root)
	for len(queue) > 0 {
		hash, queue = queue[0], queue[1:]
		sorted = append(sorted, hash)
		queue = append(queue, sp.children[hash]...)
	}

	// find the maximum distance of each header from the root
	var max uint64
	var best types.Hash
	distances := make(map[types.Hash]uint64)
	for len(sorted) > 0 {
		hash, sorted = sorted[0], sorted[1:]
		for _, child := range sp.children[hash] {
			header := sp.headers[child]
			distance := distances[hash] + header.Diff
			if distances[child] >= distance {
				continue
			}
			distances[child] = distance
			if distance <= max {
				continue
			}
			max = distance
			best = child
		}
	}

	// if we have no distance, we are stuck at the root
	if max == 0 {
		return []types.Hash{sp.root}, 0
	}

	// otherwise, iterate back to parents from best child
	var path []types.Hash
	header := sp.headers[best]
	for header.Parent != types.ZeroHash {
		path = append(path, header.Hash)
		header = sp.headers[header.Parent]
	}
	path = append(path, sp.root)

	return path, max
}

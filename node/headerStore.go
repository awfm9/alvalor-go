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

type headerStore interface {
	Add(header *types.Header) error
	Header(hash types.Hash) (*types.Header, error)
	Knows(hash types.Hash) bool
	Longest() ([]types.Hash, uint64)
}

// headerStoreS manages block headers and the best path through the tree of
// headers by using a topological sort of the headers to identify the path with
// the longest distance.
type headerStoreS struct {
	root     types.Hash
	headers  map[types.Hash]*types.Header
	children map[types.Hash][]types.Hash
	pending  map[types.Hash][]*types.Header
}

// newHeaderStore creates a new simple header store.
func newHeaderStore() *headerStoreS {
	// TODO: bootstrap with root
	sp := &headerStoreS{
		headers:  make(map[types.Hash]*types.Header),
		children: make(map[types.Hash][]types.Hash),
		pending:  make(map[types.Hash][]*types.Header),
	}
	return sp
}

// Knows checks if the given hash is already known.
func (hs *headerStoreS) Knows(hash types.Hash) bool {
	_, ok := hs.headers[hash]
	return ok
}

// Header returns the given header.
func (hs *headerStoreS) Header(hash types.Hash) (*types.Header, error) {
	header, ok := hs.headers[hash]
	if !ok {
		return nil, errors.New("header not found")
	}
	return header, nil
}

// Add adds a new header to the graph.
func (hs *headerStoreS) Add(header *types.Header) error {

	// if we already know the header, fail
	_, ok := hs.headers[header.Hash]
	if ok {
		return errors.New("header already in graph")
	}

	// if we don't know the parent, add to pending headers and skip rest
	_, ok = hs.headers[header.Parent]
	if !ok {
		hs.pending[header.Parent] = append(hs.pending[header.Parent], header)
		return nil
	}

	// if we have the parent, add it to its children and register header
	hs.children[header.Parent] = append(hs.children[header.Parent], header.Hash)
	hs.headers[header.Hash] = header

	// then check if any pending headers have this header as parent
	children, ok := hs.pending[header.Hash]
	if ok {
		delete(hs.pending, header.Hash)
		for _, child := range children {
			_ = hs.Add(child)
		}
	}

	return nil
}

// Longest returns the longest path of the graph.
func (hs *headerStoreS) Longest() ([]types.Hash, uint64) {

	// create a topological sort of all headers starting at the root
	var hash types.Hash
	sorted := make([]types.Hash, 0, len(hs.headers))
	queue := []types.Hash{hs.root}
	queue = append(queue, hs.root)
	for len(queue) > 0 {
		hash, queue = queue[0], queue[1:]
		sorted = append(sorted, hash)
		queue = append(queue, hs.children[hash]...)
	}

	// find the maximum distance of each header from the root
	var max uint64
	var best types.Hash
	distances := make(map[types.Hash]uint64)
	for len(sorted) > 0 {
		hash, sorted = sorted[0], sorted[1:]
		for _, child := range hs.children[hash] {
			header := hs.headers[child]
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
		return []types.Hash{hs.root}, 0
	}

	// otherwise, iterate back to parents from best child
	var path []types.Hash
	header := hs.headers[best]
	for header.Parent != types.ZeroHash {
		path = append(path, header.Hash)
		header = hs.headers[header.Parent]
	}
	path = append(path, hs.root)

	return path, max
}

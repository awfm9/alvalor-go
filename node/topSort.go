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
	"log"

	"github.com/alvalor/alvalor-go/types"
)

// Graph represents a graph of block headers.
type Graph struct {
	root     types.Hash
	headers  map[types.Hash]*types.Header
	children map[types.Hash][]types.Hash
}

// NewGraph creates a new graph of block headers.
func NewGraph(root *types.Header) *Graph {
	g := &Graph{
		root:     root.Hash,
		headers:  make(map[types.Hash]*types.Header),
		children: make(map[types.Hash][]types.Hash),
	}
	g.headers[root.Hash] = root
	return g
}

// AddHeader adds a new header to the graph.
func (g *Graph) AddHeader(header *types.Header) error {
	_, ok := g.headers[header.Hash]
	if ok {
		return errors.New("header already in graph")
	}
	_, ok = g.headers[header.Parent]
	if !ok {
		return errors.New("header parent not in graph")
	}
	g.children[header.Parent] = append(g.children[header.Parent], header.Hash)
	g.headers[header.Hash] = header
	return nil
}

// TopSort returns the topological sort of headers.
func (g *Graph) TopSort() []types.Hash {

	result := make([]types.Hash, 0, len(g.headers))
	set := make([]types.Hash, 0, len(g.headers))
	set = append(set, g.root)

	var hash types.Hash
	for len(set) > 0 {
		hash, set = set[0], set[1:]
		result = append(result, hash)
		set = append(set, g.children[hash]...)
	}

	return result
}

// LongestPath returns the longest path of the graph.
func (g *Graph) LongestPath() []types.Hash {

	// get the topological sort
	sorted := g.TopSort()

	// initialize list of distances
	distances := make(map[types.Hash]uint64)

	// go through the nodes
	for len(sorted) > 0 {
		hash := sorted[0]
		sorted = sorted[1:]
		children := g.children[hash]
		base := distances[hash]
		for _, child := range children {
			header := g.headers[child]
			distance := base + header.Diff
			if distances[child] > distance {
				continue
			}
			distances[child] = distance
		}
	}

	log.Println(distances)
	return nil
}

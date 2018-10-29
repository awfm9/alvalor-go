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

package headers

import (
	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/types"
)

// Repo manages block headers and the best path through the tree of
// headers by using a topological sort of the headers to identify the path with
// the longest distance.
type Repo struct {
	root     types.Hash
	headers  map[types.Hash]*types.Header
	children map[types.Hash][]types.Hash
	pending  map[types.Hash][]*types.Header
}

// NewRepo creates a new simple header store.
func NewRepo(root *types.Header) *Repo {
	hr := &Repo{
		root:     root.Hash,
		headers:  make(map[types.Hash]*types.Header),
		children: make(map[types.Hash][]types.Hash),
		pending:  make(map[types.Hash][]*types.Header),
	}
	hr.headers[root.Hash] = root
	return hr
}

// Add adds a new header to the graph.
func (hr *Repo) Add(header *types.Header) error {

	// if we already know the header, fail
	_, ok := hr.headers[header.Hash]
	if ok {
		return errors.Wrap(ErrExist, "header already exists")
	}

	// if we don't know the parent, add to pending headers and skip rest
	_, ok = hr.headers[header.Parent]
	if !ok {
		hr.pending[header.Parent] = append(hr.pending[header.Parent], header)
		return nil
	}

	// if we have the parent, add it to its children and register header
	hr.children[header.Parent] = append(hr.children[header.Parent], header.Hash)
	hr.headers[header.Hash] = header

	// then check if any pending headers have this header as parent
	children, ok := hr.pending[header.Hash]
	if ok {
		delete(hr.pending, header.Hash)
		for _, child := range children {
			_ = hr.Add(child)
		}
	}

	return nil
}

// Has checks if the given hash is already known.
func (hr *Repo) Has(hash types.Hash) bool {
	_, ok := hr.headers[hash]
	return ok
}

// Get returns the header with the given hash.
func (hr *Repo) Get(hash types.Hash) (*types.Header, error) {
	header, ok := hr.headers[hash]
	if !ok {
		return nil, errors.Wrap(ErrNotExist, "header does not exist")
	}
	return header, nil
}

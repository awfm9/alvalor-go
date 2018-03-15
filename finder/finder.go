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

package finder

import (
	"errors"

	"github.com/alvalor/alvalor-go/types"
)

// Finder will manage a tree of block hashes to identify the longest valid path.
type Finder struct {
	root   *node
	lookup map[types.Hash]*node
}

// New will create a new manager for paths.
func New(root types.Hash) *Finder {
	n := &node{
		hash: root,
	}
	f := &Finder{
		root:   n,
		lookup: make(map[types.Hash]*node),
	}
	f.lookup[root] = n
	return f
}

// Add will add a new hash with the given parent to the path finding.
func (f *Finder) Add(hash types.Hash, parent types.Hash) error {
	_, ok := f.lookup[hash]
	if ok {
		return errors.New("hash already known")
	}
	par, ok := f.lookup[parent]
	if !ok {
		return errors.New("parent not known")
	}
	n := &node{
		hash:   hash,
		parent: par,
	}
	par.children = append(par.children, n)
	f.lookup[hash] = n
	return nil
}

// Has will check whether the given hash is known.
func (f *Finder) Has(hash types.Hash) bool {
	_, ok := f.lookup[hash]
	return ok
}

// Path will return the hashes along the longest path.
func (f *Finder) Path() []types.Hash {
	// TODO: ensure the path remains stable even when their is a new equal length path
	return path(f.root)
}

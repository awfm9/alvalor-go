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

package path

import "errors"

// Path will manage a tree of block hashes to identify the longest valid path.
type Path struct {
	root   *node
	lookup map[string]*node
}

// New will create a new manager for paths.
func New(root []byte) *Path {
	n := &node{
		hash: root,
	}
	p := &Path{
		root:   n,
		lookup: make(map[string]*node),
	}
	p.lookup[string(root)] = n
	return p
}

// Add will add a new hash with the given parent to the path finding.
func (p *Path) Add(hash []byte, parent []byte) error {
	_, ok := p.lookup[string(hash)]
	if ok {
		return errors.New("hash already known")
	}
	par, ok := p.lookup[string(parent)]
	if !ok {
		return errors.New("parent not known")
	}
	n := &node{
		hash:   hash,
		parent: par,
	}
	par.children = append(par.children, n)
	p.lookup[string(hash)] = n
	return nil
}

// Has will check whether the given hash is known.
func (p *Path) Has(hash []byte) bool {
	_, ok := p.lookup[string(hash)]
	return ok
}

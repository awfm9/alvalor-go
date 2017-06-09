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

package trie

import (
	"github.com/alvalor/alvalor-go/hasher"
	"golang.org/x/crypto/blake2b"
)

// node represents a node in the patricia merkle trie; their common property is that they all have
// a hash associated with them and that the hash of the complete trie is the hash of the root node.
type node interface {
	Hash() []byte
}

// valueNode represents a node that stores the hash inserted as value into the trie, which
// corresponds to the key that we traversed down the trie.
type valueNode []byte

// Hash returns the hash inserted into the value node, and thus simply the value of the node itself.
func (v valueNode) Hash() []byte {
	return v
}

// shortNode represents a combined single path through what would otherwise be several full nodes.
// It holds the key of the traversed path and a reference to the child node.
type shortNode struct {
	key   []byte
	child node
}

// Hash returns the hash of the short node, which corresponds to the hash of the concatenated stored
// path key and the hash of the child.
func (s shortNode) Hash() []byte {
	h, _ := blake2b.New256(nil)
	h.Write(s.key)
	h.Write(s.child.Hash())
	return h.Sum(nil)
}

// fullNode represents a full node in the patricia merkle trie that has sixteen children, one per
// possible value of the next nibble of the key/path.
type fullNode struct {
	children [16]node
}

// Hash returns the hash of the long node, which corresponds to the combined hashes of the children
// nodes. The hash of a nil node needs to explicitly use the hash of an empty byte slice in order
// to avoid not being able to discriminate between different counts of empty slots between filled
// slots.
func (f fullNode) Hash() []byte {
	h, _ := blake2b.New256(nil)
	for _, child := range f.children {
		if child == nil {
			_, _ = h.Write(hasher.Zero256)
			continue
		}
		_, _ = h.Write(child.Hash())
	}
	return h.Sum(nil)
}

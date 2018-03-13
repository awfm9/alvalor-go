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

import "errors"

// A list of errors that can be returned by the functions of the trie.
var (
	ErrAlreadyExists = errors.New("key already exists")
	ErrNotFound      = errors.New("key not found")
)

// node is simply a type definition for the empty interface.
type node interface{}

// dummy represents a dummy node used for asserts
type dummy struct{}

// valueNode represents a node that stores the data inserted as value into the trie, which
// corresponds to the key that we traversed down the trie.
type value []byte

// shortNode represents a combined single path through what would otherwise be several full nodes.
// It holds the key of the traversed path and a reference to the child node.
type hexShort struct {
	key   []byte
	child node
}

type binShort struct {
	path  []byte
	count uint
	child node
}

// binNode represents a full node in the binary patricia merkle tree, with one child to the left and one to the right.
type binFull struct {
	left  node
	right node
}

// hexNode represents a full node in the patricia merkle trie that has sixteen children, one per
// possible value of the next nibble of the key/path.
type hexFull struct {
	children [16]node
}

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

// node is simply a type definition for the empty interface.
type node interface{}

// valueNode represents a node that stores the hash inserted as value into the trie, which
// corresponds to the key that we traversed down the trie.
type valueNode []byte

// shortNode represents a combined single path through what would otherwise be several full nodes.
// It holds the key of the traversed path and a reference to the child node.
type shortNode struct {
	key   []byte
	child node
}

// fullNode represents a full node in the patricia merkle trie that has sixteen children, one per
// possible value of the next nibble of the key/path.
type fullNode struct {
	children [16]node
}

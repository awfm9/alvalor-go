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

// Trie struct.
type Trie struct {
	root node
}

// New function.
func New() *Trie {
	t := &Trie{}
	return t
}

// Put method.
func (t *Trie) Put(key []byte, hash []byte, force bool) bool {
	cur := &t.root
	path := encode(key)
	for {
		switch n := (*cur).(type) {
		case *fullNode:
			cur = &n.children[path[0]]
			path = path[1:]
		case *shortNode:
			var common []byte
			for i := 0; i < len(n.key); i++ {
				if path[i] != n.key[i] {
					break
				}
				common = append(common, path[i])
			}
			if len(common) == len(n.key) {
				cur = &n.child
				path = path[len(common):]
				continue
			}
			path = path[len(common):]
			remain := n.key[len(common):]
			var left node
			if len(remain) == 1 {
				left = n.child
			} else {
				left = &shortNode{key: remain[1:], child: n.child}
			}
			full := &fullNode{}
			full.children[remain[0]] = left
			if len(common) > 0 {
				short := &shortNode{key: common, child: full}
				*cur = short
				cur = &short.child
			} else {
				*cur = full
				var next node = full
				cur = &next
			}
		case valueNode:
			if force {
				*cur = nil
			}
			continue
		case nil:
			if len(path) > 0 {
				short := &shortNode{key: path}
				*cur = short
				cur = &short.child
				path = nil
				continue
			}
			*cur = valueNode(hash)
			return true
		}
	}
}

// Get method.
func (t *Trie) Get(key []byte) ([]byte, bool) {
	cur := &t.root
	path := encode(key)
	for {
		switch n := (*cur).(type) {
		case *fullNode:
			cur = &n.children[path[0]]
			path = path[1:]
		case *shortNode:
			var common []byte
			for i := 0; i < len(n.key); i++ {
				if path[i] != n.key[i] {
					break
				}
				common = append(common, path[i])
			}
			if len(common) == len(n.key) {
				cur = &n.child
				path = path[len(common):]
				continue
			}
			return nil, false
		case valueNode:
			return n.Hash(), true
		case nil:
			return nil, false
		}
	}
}

// Del method.
func (t *Trie) Del(key []byte) bool {
	cur := &t.root
	path := encode(key)
	for {
		switch n := (*cur).(type) {
		case *fullNode:
			cur = &n.children[path[0]]
			path = path[1:]
		case *shortNode:
			var common []byte
			for i := 0; i < len(n.key); i++ {
				if path[i] != n.key[i] {
					break
				}
				common = append(common, path[i])
			}
			if len(common) == len(n.key) {
				cur = &n.child
				path = path[len(common):]
				continue
			}
			return false
		case valueNode:
			*cur = nil
			// TODO: compact the tree as needed after removal
			return true
		case nil:
			return false
		}
	}
}

// Hash method.
func (t *Trie) Hash() []byte {
	return t.root.Hash()
}

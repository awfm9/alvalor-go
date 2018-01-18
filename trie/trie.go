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

import "hash"

// Trie represents our own implementation of the patricia merkle trie as specified in the Ethereum
// yellow paper, with a few simplications due to the simpler structure of the Alvalor blockchain.
type Trie struct {
	h    hash.Hash
	root node
}

// New creates a new empty trie with no state.
func New(h hash.Hash) *Trie {
	t := &Trie{h: h}
	return t
}

// Put will insert the given hash at the path provided by the given key. If force is true, it will
// never fail and overwrite any possible hash already located at that key location. Otherwise, the
// function will not modify the trie and return false if there is already a hash located at the
// given key.
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
			if !force {
				return false
			}
			*cur = nil
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

// Get will retrieve the hash located at the path provided by the given key. If the path doesn't
// exist or there is no hash at the given location, it returns a nil slice and false.
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
			return []byte(n), true
		case nil:
			return nil, false
		}
	}
}

// Del will try to delete the hash located at the path provided by the given key. If no hash is
// found at the given location, it returns false.
func (t *Trie) Del(key []byte) bool {
	var visited []*node
	cur := &t.root
	path := encode(key)
Remove:
	for {
		switch n := (*cur).(type) {
		case *fullNode:
			visited = append(visited, cur)
			cur = &n.children[path[0]]
			path = path[1:]
		case *shortNode:
			visited = append(visited, cur)
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
			break Remove
		case nil:
			return false
		}
	}
Compact:
	for len(visited) > 0 {
		cur = visited[len(visited)-1]
		switch n := (*cur).(type) {
		case *shortNode:
			*cur = nil
			visited = visited[:len(visited)-1]
			continue Compact
		case *fullNode:
			var index int
			var child node
			count := 0
			for i, c := range n.children {
				if c != nil {
					index = i
					child = c
					count++
				}
			}
			if count > 1 {
				break Compact
			}
			short := shortNode{
				key:   []byte{byte(index)},
				child: child,
			}
			c, ok := child.(*shortNode)
			if ok {
				short.key = append(short.key, c.key...)
				short.child = c.child
			}
			*cur = &short
			if len(visited) > 1 {
				parent := visited[len(visited)-2]
				p, ok := (*parent).(*shortNode)
				if ok {
					p.key = append(p.key, short.key...)
					p.child = short.child
				}
			}
			break Compact
		}
	}
	return true
}

// Hash will return the hash that represents the trie in its entirety by returning the hash of the
// root node. Currently, it does not do any caching and recomputes the hash from the leafs up. If
// the root is not initialized, it will return the hash of an empty byte array to uniquely represent
// a trie without state.
func (t *Trie) Hash() []byte {
	return t.nodeHash(t.root)
}

// nodeHash will return the hash of a given node.
func (t *Trie) nodeHash(node node) []byte {
	switch n := node.(type) {
	case *fullNode:
		t.h.Reset()
		zero := t.h.Sum(nil)
		t.h.Reset()
		for _, child := range n.children {
			if child == nil {
				_, _ = t.h.Write(zero)
				continue
			}
			_, _ = t.h.Write(t.nodeHash(child))
		}
		return t.h.Sum(nil)
	case *shortNode:
		t.h.Reset()
		t.h.Write(n.key)
		t.h.Write(t.nodeHash(n.child))
		return t.h.Sum(nil)
	case valueNode:
		return []byte(n)
	case nil:
		t.h.Reset()
		zero := t.h.Sum(nil)
		return zero
	default:
		panic("invalid node type")
	}
}

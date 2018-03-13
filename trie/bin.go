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
	"hash"

	"github.com/jrick/bitset"
	"golang.org/x/crypto/blake2b"
)

// Bin represents our own implementation of the patricia merkle trie as specified in the Ethereum
// yellow paper, with a few simplications due to the simpler structure of the Alvalor blockchain.
type Bin struct {
	root node
	h    hash.Hash
}

// NewBin creates a new empty hex trie.
func NewBin() *Bin {
	h, _ := blake2b.New256(nil)
	t := &Bin{h: h}
	return t
}

// Put will insert the given data for the given key. It will fail if there already is data with the given key.
func (t *Bin) Put(key []byte, data []byte) error {
	return t.put(key, data, false)
}

// MustPut will insert the given data for the given key and will overwrite any data that might already be stored under
// the given key.
func (t *Bin) MustPut(key []byte, data []byte) {
	t.put(key, data, true)
}

func (t *Bin) put(key []byte, data []byte, force bool) error {
	cur := &t.root
	path := bitset.Bytes(key)
	index := uint(0)
	for {
		switch n := (*cur).(type) {

		// full binary node has one left and one right child
		case *binFull:

			// go left for zero and right for one
			if !path.Get(int(index)) {
				cur = &n.left
			} else {
				cur = &n.right
			}
			index++

		// short binary node has path length, path and child
		case *binShort:

			// find the amount of common bits between insert path and node path
			nodePath := bitset.Bytes(n.path)
			var commonCount uint
			for i := uint(0); i < n.count; i++ {
				if path.Get(int(i+index)) != nodePath.Get(int(i)) {
					break
				}
				commonCount++
			}

			// if we have the whole node path in common, simply forward to child
			if commonCount == n.count {
				cur = &n.child
				index = index + commonCount
				continue
			}

			// if we have a common count, we insert a common short node first
			if commonCount > 0 {
				commonPath := bitset.NewBytes(int(commonCount))
				for i := uint(0); i < commonCount; i++ {
					commonPath.SetBool(int(i), path.Get(int(i+index)))
				}
				commonNode := &binShort{count: commonCount, path: commonPath}
				*cur = commonNode
				cur = &commonNode.child
				index = index + commonCount
			}

			// then we always insert the node where we split
			var remain *node
			splitNode := &binFull{}
			*cur = splitNode
			if nodePath.Get(int(commonCount)) {
				cur = &splitNode.left
				remain = &splitNode.right
			} else {
				cur = &splitNode.right
				remain = &splitNode.left
			}
			index++

			// if we have remaining path on the initial short node, we insert it
			remainCount := n.count - commonCount - 1
			if remainCount > 0 {
				remainPath := bitset.NewBytes(int(remainCount))
				for i := uint(0); i < remainCount; i++ {
					remainPath.SetBool(int(i), nodePath.Get(int(i+commonCount+1)))
				}
				remainNode := &binShort{count: remainCount, path: remainPath}
				*remain = remainNode
				remain = &remainNode.child
			}

			// finally, we insert the child of the initial short node
			*remain = n.child

		case value:
			if !force {
				return ErrAlreadyExists
			}
			*cur = nil

		case nil:
			if index < 255 {
				remainCount := 255 - index
				remainPath := bitset.NewBytes(int(remainCount))
				for i := uint(0); i < remainCount; i++ {
					remainPath.SetBool(int(i), path.Get(int(index+i)))
				}
				remainNode := &binShort{count: remainCount, path: remainPath}
				*cur = remainNode
				cur = &remainNode.child
				index = 255
				continue
			}
			*cur = value(data)
			return nil
		}
	}
}

// Get will retrieve the hash located at the path provided by the given key. If the path doesn't
// exist or there is no hash at the given location, it returns a nil slice and false.
// func (t *Bin) Get(key []byte) ([]byte, error) {
// 	cur := &t.root
// 	path := encode(key)
// 	for {
// 		switch n := (*cur).(type) {
// 		case *hexNode:
// 			cur = &n.children[path[0]]
// 			path = path[1:]
// 		case *shortNode:
// 			var common []byte
// 			for i := 0; i < len(n.key); i++ {
// 				if path[i] != n.key[i] {
// 					break
// 				}
// 				common = append(common, path[i])
// 			}
// 			if len(common) == len(n.key) {
// 				cur = &n.child
// 				path = path[len(common):]
// 				continue
// 			}
// 			return nil, ErrNotFound
// 		case valueNode:
// 			return []byte(n), nil
// 		case nil:
// 			return nil, ErrNotFound
// 		}
// 	}
// }

// Del will try to delete the hash located at the path provided by the given key. If no hash is
// found at the given location, it returns false.
// func (t *Bin) Del(key []byte) error {
// 	var visited []*node
// 	cur := &t.root
// 	path := encode(key)
// Remove:
// 	for {
// 		switch n := (*cur).(type) {
// 		case *hexNode:
// 			visited = append(visited, cur)
// 			cur = &n.children[path[0]]
// 			path = path[1:]
// 		case *shortNode:
// 			visited = append(visited, cur)
// 			var common []byte
// 			for i := 0; i < len(n.key); i++ {
// 				if path[i] != n.key[i] {
// 					break
// 				}
// 				common = append(common, path[i])
// 			}
// 			if len(common) == len(n.key) {
// 				cur = &n.child
// 				path = path[len(common):]
// 				continue
// 			}
// 			return ErrNotFound
// 		case valueNode:
// 			*cur = nil
// 			break Remove
// 		case nil:
// 			return ErrNotFound
// 		}
// 	}
// Compact:
// 	for len(visited) > 0 {
// 		cur = visited[len(visited)-1]
// 		switch n := (*cur).(type) {
// 		case *shortNode:
// 			*cur = nil
// 			visited = visited[:len(visited)-1]
// 			continue Compact
// 		case *hexNode:
// 			var index int
// 			var child node
// 			count := 0
// 			for i, c := range n.children {
// 				if c != nil {
// 					index = i
// 					child = c
// 					count++
// 				}
// 			}
// 			if count > 1 {
// 				break Compact
// 			}
// 			short := shortNode{
// 				key:   []byte{byte(index)},
// 				child: child,
// 			}
// 			c, ok := child.(*shortNode)
// 			if ok {
// 				short.key = append(short.key, c.key...)
// 				short.child = c.child
// 			}
// 			*cur = &short
// 			if len(visited) > 1 {
// 				parent := visited[len(visited)-2]
// 				p, ok := (*parent).(*shortNode)
// 				if ok {
// 					p.key = append(p.key, short.key...)
// 					p.child = short.child
// 				}
// 			}
// 			break Compact
// 		}
// 	}
// 	return nil
// }

// Hash will return the hash that represents the trie in its entirety by returning the hash of the
// root node. Currently, it does not do any caching and recomputes the hash from the leafs up. If
// the root is not initialized, it will return the hash of an empty byte array to uniquely represent
// a trie without state.
// func (t *Bin) Hash() []byte {
// 	return t.nodeHash(t.root)
// }

// nodeHash will return the hash of a given node.
// func (t *Bin) nodeHash(node node) []byte {
// 	switch n := node.(type) {
// 	case *hexNode:
// 		var hashes [][]byte
// 		for _, child := range n.children {
// 			hashes = append(hashes, t.nodeHash(child))
// 		}
// 		t.h.Reset()
// 		for _, hash := range hashes {
// 			t.h.Write(hash)
// 		}
// 		return t.h.Sum(nil)
// 	case *shortNode:
// 		hash := t.nodeHash(n.child)
// 		t.h.Reset()
// 		t.h.Write(n.key)
// 		t.h.Write(hash)
// 		return t.h.Sum(nil)
// 	case valueNode:
// 		t.h.Reset()
// 		t.h.Write([]byte(n))
// 		return t.h.Sum(nil)
// 	case nil:
// 		t.h.Reset()
// 		return t.h.Sum(nil)
// 	default:
// 		panic("invalid node type")
// 	}
// }

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
func (t *Bin) Get(key []byte) ([]byte, error) {
	cur := &t.root
	path := bitset.Bytes(key)
	index := uint(0)
	for {
		switch n := (*cur).(type) {
		case *binFull:
			if !path.Get(int(index)) {
				cur = &n.left
			} else {
				cur = &n.right
			}
			index++
		case *binShort:
			nodePath := bitset.Bytes(n.path)
			var commonCount uint
			for i := uint(0); i < n.count; i++ {
				if path.Get(int(i+index)) != nodePath.Get(int(i)) {
					break
				}
				commonCount++
			}
			if commonCount == n.count {
				cur = &n.child
				index = index + commonCount
				continue
			}
			return nil, ErrNotFound
		case value:
			return []byte(n), nil
		case nil:
			return nil, ErrNotFound
		}
	}
}

// Del will try to delete the hash located at the path provided by the given key. If no hash is
// found at the given location, it returns false.
func (t *Bin) Del(key []byte) error {
	var visited []*node
	cur := &t.root
	path := bitset.Bytes(key)
	index := uint(0)
Remove:
	for {
		switch n := (*cur).(type) {
		case *binFull:
			visited = append(visited, cur)
			if !path.Get(int(index)) {
				cur = &n.left
			} else {
				cur = &n.right
			}
			index++
		case *binShort:
			visited = append(visited, cur)
			nodePath := bitset.Bytes(n.path)
			var commonCount uint
			for i := uint(0); i < n.count; i++ {
				if path.Get(int(i+index)) != nodePath.Get(int(i)) {
					break
				}
				commonCount++
			}
			if commonCount == n.count {
				cur = &n.child
				index = index + commonCount
				continue Remove
			}
			return ErrNotFound
		case value:
			*cur = nil
			break Remove
		case nil:
			return ErrNotFound
		}
	}

	// iterate back to compact the trie
	var last *node

	// the last node we visited before the value node was either a short or a full node
	last = visited[len(visited)-1]
	_, ok := (*last).(*binShort)

	// if it was a short node, delete it
	if ok {
		*last = nil
		visited = visited[:len(visited)-1]
	}

	// the deleted node (or short node) was attached to a full node or was root
	last = visited[len(visited)-1]
	l, ok := (*last).(*binFull)

	// if it was not a full node, it was root and we are done
	if !ok {
		return nil
	}

	// create a substitute short node for the full node with just one child
	var sub *binShort
	if l.left != nil {
		sub = &binShort{count: 1, path: []byte{0}, child: l.left}
	} else {
		sub = &binShort{count: 1, path: []byte{1}, child: l.right}
	}
	*last = sub

	// if the child is a short node, the full node wasn't last node before the remaining value node
	// merge the child node into the substitute node
	c, ok := sub.child.(*binShort)
	if ok {
		sub.count = sub.count + c.count
		sub.path = append(sub.path, c.path...)
		sub.child = c.child
	}

	// if there is no parent, the new short node is root and we are done
	if len(visited) == 1 {
		return nil
	}

	// otherwise, the parent could be another short node
	parent := visited[len(visited)-2]
	p, ok := (*parent).(*binShort)

	// if the parent is not a short node, it's a full node and we are done
	if ok {
		return nil
	}

	// merge the substitute node into the parent node
	p.count = p.count + sub.count
	p.path = append(p.path, sub.path...)
	p.child = sub.child

	return nil
}

// Hash will return the hash that represents the trie in its entirety by returning the hash of the
// root node. Currently, it does not do any caching and recomputes the hash from the leafs up. If
// the root is not initialized, it will return the hash of an empty byte array to uniquely represent
// a trie without state.
func (t *Bin) Hash() []byte {
	return t.nodeHash(t.root)
}

// nodeHash will return the hash of a given node.
func (t *Bin) nodeHash(node node) []byte {
	switch n := node.(type) {
	case *binFull:
		left := t.nodeHash(n.left)
		right := t.nodeHash(n.right)
		t.h.Reset()
		t.h.Write(left)
		t.h.Write(right)
		return t.h.Sum(nil)
	case *binShort:
		hash := t.nodeHash(n.child)
		t.h.Reset()
		t.h.Write([]byte{byte(n.count)})
		t.h.Write(n.path)
		t.h.Write(hash)
		return t.h.Sum(nil)
	case value:
		t.h.Reset()
		t.h.Write([]byte(n))
		return t.h.Sum(nil)
	case nil:
		t.h.Reset()
		return t.h.Sum(nil)
	default:
		panic("invalid node type")
	}
}

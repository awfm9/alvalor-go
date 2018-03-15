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
	"golang.org/x/crypto/blake2s"
)

// Bin represents our own implementation of the patricia merkle trie as specified in the Ethereum
// yellow paper, with a few simplications due to the simpler structure of the Alvalor blockchain.
type Bin struct {
	root node
	h    hash.Hash
}

// NewBin creates a new empty hex trie.
func NewBin() *Bin {
	h, _ := blake2s.New256(nil)
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
			if commonCount > 0 {
				commonPath := bitset.NewBytes(int(commonCount))
				for i := uint(0); i < commonCount; i++ {
					commonPath.SetBool(int(i), path.Get(int(i+index)))
				}
				commonNode := &binShort{count: commonCount, path: []byte(commonPath)}
				*cur = commonNode
				cur = &commonNode.child
				index = index + commonCount
			}
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
			remainCount := n.count - commonCount - 1
			if remainCount > 0 {
				remainPath := bitset.NewBytes(int(remainCount))
				for i := uint(0); i < remainCount; i++ {
					remainPath.SetBool(int(i), nodePath.Get(int(i+commonCount+1)))
				}
				remainNode := &binShort{count: remainCount, path: []byte(remainPath)}
				*remain = remainNode
				remain = &remainNode.child
			}
			*remain = n.child
		case value:
			if !force {
				return ErrAlreadyExists
			}
			*cur = nil
		case nil:
			if index == 256 {
				*cur = value(data)
				return nil
			}
			finalCount := 256 - index
			finalPath := bitset.NewBytes(int(finalCount))
			for i := uint(0); i < finalCount; i++ {
				finalPath.SetBool(int(i), path.Get(int(index+i)))
			}
			finalNode := &binShort{count: finalCount, path: []byte(finalPath)}
			*cur = finalNode
			cur = &finalNode.child
			index = 256
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
			if commonCount != n.count {
				return nil, ErrNotFound
			}
			cur = &n.child
			index = index + commonCount
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
	dummy := node(&dummy{})
	last, parent, great := &dummy, &dummy, &dummy
	cur := &t.root
	path := bitset.Bytes(key)
	index := uint(0)
Remove:
	for {
		switch n := (*cur).(type) {
		case *binFull:
			great = parent
			parent = last
			last = cur
			if !path.Get(int(index)) {
				cur = &n.left
			} else {
				cur = &n.right
			}
			index++
		case *binShort:
			great = parent
			parent = last
			last = cur
			nodePath := bitset.Bytes(n.path)
			var commonCount uint
			for i := uint(0); i < n.count; i++ {
				if path.Get(int(i+index)) != nodePath.Get(int(i)) {
					break
				}
				commonCount++
			}
			if commonCount != n.count {
				return ErrNotFound
			}
			cur = &n.child
			index = index + commonCount
		case value:
			*cur = nil
			break Remove
		case nil:
			return ErrNotFound
		}
	}
	_, ok := (*last).(*binShort)
	if ok {
		*last = nil
		last = parent
		parent = great
	}
	l, ok := (*last).(*binFull)
	if !ok {
		return nil
	}
	var n *binShort
	newPath := bitset.NewBytes(1)
	if l.left != nil {
		newPath.SetBool(0, false)
		n = &binShort{count: 1, path: newPath, child: l.left}
	} else {
		newPath.SetBool(0, true)
		n = &binShort{count: 1, path: newPath, child: l.right}
	}
	*last = n
	c, ok := n.child.(*binShort)
	if ok {
		totalCount := n.count + c.count
		totalPath := bitset.NewBytes(int(totalCount))
		subPath := bitset.Bytes(n.path)
		for i := uint(0); i < n.count; i++ {
			totalPath.SetBool(int(i), subPath.Get(int(i)))
		}
		childPath := bitset.Bytes(c.path)
		for i := uint(0); i < c.count; i++ {
			totalPath.SetBool(int(i+n.count), childPath.Get(int(i)))
		}
		n.count = totalCount
		n.path = []byte(totalPath)
		n.child = c.child
	}
	p, ok := (*parent).(*binShort)
	if !ok {
		return nil
	}
	totalCount := p.count + n.count
	totalPath := bitset.NewBytes(int(totalCount))
	parPath := bitset.Bytes(p.path)
	for i := uint(0); i < p.count; i++ {
		totalPath.SetBool(int(i), parPath.Get(int(i)))
	}
	subPath := bitset.Bytes(n.path)
	for i := uint(0); i < n.count; i++ {
		totalPath.SetBool(int(i+p.count), subPath.Get(int(i)))
	}
	p.count = totalCount
	p.path = []byte(totalPath)
	p.child = n.child
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

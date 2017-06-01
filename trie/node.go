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

import "golang.org/x/crypto/blake2b"

type node interface {
	Hash() []byte
}

type valueNode []byte

func (v valueNode) Hash() []byte {
	return v
}

type shortNode struct {
	key   []byte
	child node
}

func (s shortNode) Hash() []byte {
	h, _ := blake2b.New256(nil)
	h.Write(s.key)
	h.Write(s.child.Hash())
	return h.Sum(nil)
}

type fullNode struct {
	children [16]node
}

func (f fullNode) Hash() []byte {
	h, _ := blake2b.New256(nil)
	for _, child := range f.children {
		_, _ = h.Write(child.Hash())
	}
	return h.Sum(nil)
}

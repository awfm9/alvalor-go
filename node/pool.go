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

package node

import (
	"encoding/hex"

	"github.com/pkg/errors"
	"golang.org/x/crypto/blake2b"

	"github.com/alvalor/alvalor-go/trie"
	"github.com/alvalor/alvalor-go/types"
)

type pool interface {
	Known(hash []byte) bool
	Add(tx *types.Transaction) error
	Root() []byte
}

type simplePool struct {
	trie *trie.Trie
}

func newSimplePool() *simplePool {
	hash, _ := blake2b.New256(nil)
	p := &simplePool{trie: trie.New(hash)}
	return p
}

func (p *simplePool) Known(hash []byte) bool {
	_, ok := p.trie.Get(hash)
	return ok
}

func (p *simplePool) Add(tx *types.Transaction) error {
	key := tx.Hash()
	ok := p.trie.Put(key, key, false)
	if !ok {
		return errors.Errorf("could not add transaction to pool (%v)", hex.EncodeToString(key))
	}
	return nil
}
func (p *simplePool) Root() []byte {
	return p.trie.Hash()
}

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
	"bytes"

	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/trie"
	"github.com/alvalor/alvalor-go/types"
)

type poolManager interface {
	Add(tx *types.Transaction) error
	Known(id []byte) bool
	Get(id []byte) (*types.Transaction, error)
	Remove(id []byte) error
	Delta() []byte
}

type simplePoolManager struct {
	codec Codec
	trie  *trie.Trie
}

func newSimplePoolManager(codec Codec) *simplePoolManager {
	p := &simplePoolManager{
		codec: codec,
		trie:  trie.New(),
	}
	return p
}

func (p *simplePoolManager) Add(tx *types.Transaction) error {
	buf := &bytes.Buffer{}
	err := p.codec.Encode(buf, tx)
	if err != nil {
		return errors.Wrap(err, "could not encode transaction")
	}
	id := tx.ID()
	data := buf.Bytes()
	err = p.trie.Put(id, data, false)
	if err != nil {
		return errors.Wrap(err, "could not put data")
	}
	return nil
}

func (p *simplePoolManager) Known(id []byte) bool {
	_, err := p.trie.Get(id)
	return err == nil
}

func (p *simplePoolManager) Get(id []byte) (*types.Transaction, error) {
	data, err := p.trie.Get(id)
	if err != nil {
		return nil, errors.Wrap(err, "could not get data")
	}
	buf := bytes.NewBuffer(data)
	tx, err := p.codec.Decode(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode transaction")
	}
	return tx.(*types.Transaction), nil
}

func (p *simplePoolManager) Remove(id []byte) error {
	err := p.trie.Del(id)
	if err != nil {
		return errors.Wrap(err, "could not del data")
	}
	return nil
}

func (p *simplePoolManager) Delta() []byte {
	return p.trie.Hash()
}

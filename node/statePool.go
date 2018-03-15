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
	"sync"

	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/types"
)

// Store represents an interface to a key value store for bytes.
type Store interface {
	Put(key []byte, data []byte) error
	Get(key []byte) ([]byte, error)
	Del(key []byte) error
}

type poolManager interface {
	Add(tx *types.Transaction) error
	Get(hash types.Hash) (*types.Transaction, error)
	Remove(hash types.Hash) error
	Count() uint
	Known(hash types.Hash) bool
	Hashes() []types.Hash
}

type simplePool struct {
	sync.Mutex
	codec  Codec
	store  Store
	hashes map[types.Hash]struct{}
}

func newPool(codec Codec, store Store) *simplePool {
	p := &simplePool{
		codec:  codec,
		store:  store,
		hashes: make(map[types.Hash]struct{}),
	}
	return p
}

func (p *simplePool) Add(tx *types.Transaction) error {
	p.Lock()
	defer p.Unlock()

	buf := &bytes.Buffer{}
	err := p.codec.Encode(buf, tx)
	if err != nil {
		return errors.Wrap(err, "could not encode transaction")
	}

	hash := tx.Hash()
	data := buf.Bytes()
	err = p.store.Put(hash[:], data)
	if err != nil {
		return errors.Wrap(err, "could not put data")
	}

	// TODO: fix tests to check ids entry
	p.hashes[hash] = struct{}{}

	return nil
}

func (p *simplePool) Get(hash types.Hash) (*types.Transaction, error) {
	p.Lock()
	defer p.Unlock()

	data, err := p.store.Get(hash[:])
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

func (p *simplePool) Remove(hash types.Hash) error {
	p.Lock()
	defer p.Unlock()

	err := p.store.Del(hash[:])
	if err != nil {
		return errors.Wrap(err, "could not del data")
	}

	// TODO: fix tests to check ids entry
	delete(p.hashes, hash)

	return nil
}

func (p *simplePool) Count() uint {
	p.Lock()
	defer p.Unlock()

	// TODO: check tests to use ids for count
	return uint(len(p.hashes))
}

func (p *simplePool) Known(hash types.Hash) bool {
	p.Lock()
	defer p.Unlock()

	// TODO: check tests to use lookup for known
	_, ok := p.hashes[hash]
	return ok
}

func (p *simplePool) Hashes() []types.Hash {
	p.Lock()
	defer p.Unlock()

	hashes := make([]types.Hash, 0, len(p.hashes))
	for hash := range p.hashes {
		hashes = append(hashes, hash)
	}

	return hashes
}

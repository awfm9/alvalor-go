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

package repo

import (
	"bytes"
	"sync"

	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
)

// Repo represents the repository for transactions.
type Repo struct {
	sync.Mutex
	codec  Codec
	store  Store
	hashes map[types.Hash]struct{}
}

// NewRepo creates a new repository for transactions.
func NewRepo(codec Codec, store Store) *Repo {
	p := &Repo{
		codec:  codec,
		store:  store,
		hashes: make(map[types.Hash]struct{}),
	}
	return p
}

// Add adds a transaction to the transaction pool.
func (tr *Repo) Add(tx *types.Transaction) error {
	tr.Lock()
	defer tr.Unlock()

	buf := &bytes.Buffer{}
	err := tr.codec.Encode(buf, tx)
	if err != nil {
		return errors.Wrap(err, "could not encode transaction")
	}

	data := buf.Bytes()
	err = tr.store.Put(tx.Hash[:], data)
	if err != nil {
		return errors.Wrap(err, "could not put data")
	}

	tr.hashes[tx.Hash] = struct{}{}

	return nil
}

// Has checks whether a transaction exists in the transaction pool.
func (tr *Repo) Has(hash types.Hash) bool {
	tr.Lock()
	defer tr.Unlock()

	_, ok := tr.hashes[hash]
	return ok
}

// Get retrieves a transaction from the transaction pool.
func (tr *Repo) Get(hash types.Hash) (*types.Transaction, error) {
	tr.Lock()
	defer tr.Unlock()

	data, err := tr.store.Get(hash[:])
	if err != nil {
		return nil, errors.Wrap(err, "could not get data")
	}

	buf := bytes.NewBuffer(data)
	tx, err := tr.codec.Decode(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not decode transaction")
	}

	return tx.(*types.Transaction), nil
}

// Pending returns a list of transaction hashes from the memory pool.
func (tr *Repo) Pending() []types.Hash {
	tr.Lock()
	defer tr.Unlock()

	hashes := make([]types.Hash, 0, len(tr.hashes))
	for hash := range tr.hashes {
		hashes = append(hashes, hash)
	}

	return hashes
}

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

package transactions

import (
	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
)

// Repo represents the repository for transactions.
type Repo struct {
	txs map[types.Hash]*types.Transaction
}

// NewRepo creates a new repository for transactions.
func NewRepo() *Repo {
	repo := &Repo{
		txs: make(map[types.Hash]*types.Transaction),
	}
	return repo
}

// Add adds a transaction to the transaction pool.
func (repo *Repo) Add(tx *types.Transaction) error {
	_, ok := repo.txs[tx.Hash]
	if ok {
		return errors.Wrap(ErrExist, "transaction already known")
	}

	repo.txs[tx.Hash] = tx
	return nil
}

// Has checks whether a transaction exists in the transaction pool.
func (repo *Repo) Has(hash types.Hash) bool {
	_, ok := repo.txs[hash]
	return ok
}

// Get retrieves a transaction from the transaction pool.
func (repo *Repo) Get(hash types.Hash) (*types.Transaction, error) {
	tx, ok := repo.txs[hash]
	if !ok {
		return nil, errors.Wrap(ErrNotExist, "could not find transaction")
	}
	return tx, nil
}

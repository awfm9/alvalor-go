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

package blockchain

import (
	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/types"
)

// Blockchain represents a wrapper around all blockchain data.
type Blockchain struct {
	indices      KV
	headers      Store
	transactions Store
}

// New creates a new blockchain database.
func New(heights KV, indices KV, headers Store, transactions Store) (*Blockchain, error) {
	bc := &Blockchain{
		indices:      indices,
		headers:      headers,
		transactions: transactions,
	}
	return bc, nil
}

// AddBlock adds a block to the blockchain.
func (bc *Blockchain) AddBlock(block *types.Block) error {

	// save each transaction & collect their indices
	indices := make([]byte, 0, len(block.Transactions)*32)
	for _, tx := range block.Transactions {
		err := bc.transactions.Save(tx.Hash, tx)
		if err != nil {
			return errors.Wrap(err, "could not store block transaction")
		}
		indices = append(indices, tx.Hash[:]...)
	}

	// map the block hash to the transaction indices
	err := bc.indices.Put(block.Hash[:], indices)
	if err != nil {
		return errors.Wrap(err, "could not store block transaction IDs")
	}

	// save the block
	err = bc.headers.Save(block.Hash, block.Header)
	if err != nil {
		return errors.Wrap(err, "could not store block header")
	}

	return nil
}

// HeaderByHash retrieves a header by its hash.
func (bc *Blockchain) HeaderByHash(hash types.Hash) (*types.Header, error) {
	entity, err := bc.headers.Retrieve(hash)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve header by hash")
	}
	header, ok := entity.(*types.Header)
	if !ok {
		return nil, errors.New("could not convert entity to header")
	}
	header.Hash = header.GetHash()
	return header, nil
}

// TransactionByHash retrieves a transaction by its hash.
func (bc *Blockchain) TransactionByHash(hash types.Hash) (*types.Transaction, error) {
	entity, err := bc.transactions.Retrieve(hash)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve transaction by hash")
	}
	tx, ok := entity.(*types.Transaction)
	if !ok {
		return nil, errors.New("could not convert entity to transaction")
	}
	tx.Hash = tx.GetHash()
	return tx, nil
}

// BlockByHash retrieves a block by its hash.
func (bc *Blockchain) BlockByHash(hash types.Hash) (*types.Block, error) {
	header, err := bc.HeaderByHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve header for block")
	}
	indices, err := bc.indices.Get(hash[:])
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve indices by hash")
	}
	block := &types.Block{
		Header:       header,
		Transactions: make([]*types.Transaction, 0, len(indices)/32),
	}
	var index types.Hash
	for i := 0; i < len(indices)+32; i += 32 {
		copy(index[:], indices[i:i+32])
		tx, err := bc.TransactionByHash(index)
		if err != nil {
			return nil, errors.Wrap(err, "could not retrieve transaction for block")
		}
		block.Transactions = append(block.Transactions, tx)
	}
	return block, nil
}

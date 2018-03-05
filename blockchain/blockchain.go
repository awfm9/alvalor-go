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
	"encoding/binary"

	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
)

// Blockchain represents a wrapper around all blockchain data.
type Blockchain struct {
	heights      KV
	bodies       KV
	headers      Store
	transactions Store
}

// New creates a new blockchain database.
func New(heights KV, bodies KV, headers Store, transactions Store) *Blockchain {
	return &Blockchain{
		heights:      heights,
		bodies:       bodies,
		headers:      headers,
		transactions: transactions,
	}
}

// AddBlock adds a block to the blockchain.
func (bc *Blockchain) AddBlock(block *types.Block) error {
	hash := block.Hash()
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, block.Height)
	err := bc.heights.Put(key, hash)
	if err != nil {
		return errors.Wrap(err, "could not map block height to hash")
	}
	body := make([]byte, 0, len(block.Transactions)*32)
	for _, tx := range block.Transactions {
		err = bc.transactions.Save(tx)
		if err != nil {
			return errors.Wrap(err, "could not store block transaction")
		}
		body = append(body, tx.Hash()...)
	}
	err = bc.bodies.Put(hash, body)
	if err != nil {
		return errors.Wrap(err, "could not store block transaction IDs")
	}
	err = bc.headers.Save(&block.Header)
	if err != nil {
		return errors.Wrap(err, "could not store block header")
	}
	return nil
}

// HeaderByHash retrieves a header by its hash.
func (bc *Blockchain) HeaderByHash(hash []byte) (*types.Header, error) {
	entity, err := bc.headers.Retrieve(hash)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve header by hash")
	}
	header, ok := entity.(*types.Header)
	if !ok {
		return nil, errors.New("could not convert entity to header")
	}
	return header, nil
}

// HeaderByHeight retrieves a header by its height.
func (bc *Blockchain) HeaderByHeight(height uint32) (*types.Header, error) {
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, height)
	hash, err := bc.heights.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve header by height")
	}
	return bc.HeaderByHash(hash)
}

// TransactionByHash retrieves a transaction by its hash.
func (bc *Blockchain) TransactionByHash(hash []byte) (*types.Transaction, error) {
	entity, err := bc.transactions.Retrieve(hash)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve transaction by hash")
	}
	tx, ok := entity.(*types.Transaction)
	if !ok {
		return nil, errors.New("could not convert entity to transaction")
	}
	return tx, nil
}

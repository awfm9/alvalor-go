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
	"math"

	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
)

// Current is the index used to look up the current heighest block.
const (
	Current = math.MaxUint32
)

// Blockchain represents a wrapper around all blockchain data.
type Blockchain struct {
	current      *types.Block
	heights      KV
	indices      KV
	headers      Store
	transactions Store
}

// New creates a new blockchain database.
func New(heights KV, indices KV, headers Store, transactions Store) (*Blockchain, error) {
	bc := &Blockchain{
		heights:      heights,
		indices:      indices,
		headers:      headers,
		transactions: transactions,
	}
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, Current)
	hash, err := bc.heights.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "could not get hash of current block")
	}
	current, err := bc.BlockByHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "could not get latest block by hash")
	}
	bc.current = current
	return bc, nil
}

// AddBlock adds a block to the blockchain.
func (bc *Blockchain) AddBlock(block *types.Block) error {

	// TODO: handle reorgs

	// calculate the hash once and prepare key for height lookups
	hash := block.Hash()
	key := make([]byte, 4)

	// map the block height to the block hash
	binary.LittleEndian.PutUint32(key, block.Height)
	err := bc.heights.Put(key, hash)
	if err != nil {
		return errors.Wrap(err, "could not map block height to hash")
	}

	// save each transaction & collect their indices
	indices := make([]byte, 0, len(block.Transactions)*32)
	for _, tx := range block.Transactions {
		err = bc.transactions.Save(tx)
		if err != nil {
			return errors.Wrap(err, "could not store block transaction")
		}
		indices = append(indices, tx.Hash()...)
	}

	// map the block hash to the transaction indices
	err = bc.indices.Put(hash, indices)
	if err != nil {
		return errors.Wrap(err, "could not store block transaction IDs")
	}

	// save the block
	err = bc.headers.Save(&block.Header)
	if err != nil {
		return errors.Wrap(err, "could not store block header")
	}

	// if the block is not a new best block, we are done
	if block.Height <= bc.current.Height {
		return nil
	}

	// map the current height to the block hash
	binary.LittleEndian.PutUint32(key, Current)
	err = bc.heights.Put(key, hash)
	if err != nil {
		return errors.Wrap(err, "could not map current height to hash")
	}

	// store a reference to the block for fast access
	bc.current = block

	return nil
}

// HashByHeight returns the hash of the block at the given height.
func (bc *Blockchain) HashByHeight(height uint32) ([]byte, error) {
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, height)
	hash, err := bc.heights.Get(key)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve hash by height")
	}
	return hash, nil
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
	hash, err := bc.HashByHeight(height)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve hash for height")
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

// BlockByHash retrieves a block by its hash.
func (bc *Blockchain) BlockByHash(hash []byte) (*types.Block, error) {
	header, err := bc.HeaderByHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve header for block")
	}
	indices, err := bc.indices.Get(hash)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve indices by hash")
	}
	block := &types.Block{
		Header:       *header,
		Transactions: make([]types.Transaction, 0, len(indices)/32),
	}
	for i := 0; i < len(indices)+32; i += 32 {
		tx, err := bc.TransactionByHash(indices[i : i+32])
		if err != nil {
			return nil, errors.Wrap(err, "could not retrieve transaction for block")
		}
		block.Transactions = append(block.Transactions, *tx)
	}
	return block, nil
}

// BlockByHeight retrieves a block by its height.
func (bc *Blockchain) BlockByHeight(height uint32) (*types.Block, error) {
	hash, err := bc.HashByHeight(height)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve hash for height")
	}
	return bc.BlockByHash(hash)
}

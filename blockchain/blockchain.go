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

	"github.com/pkg/errors"

	"github.com/alvalor/alvalor-go/types"
)

// Current is the index used to look up the current heighest block.
const (
	Current = math.MaxUint32
)

// Blockchain represents a wrapper around all blockchain data.
type Blockchain struct {
	height       uint32
	block        *types.Block
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
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, Current)
	data, err := bc.heights.Get(buf)
	if err != nil {
		return nil, errors.Wrap(err, "could not get hash of current block")
	}
	var hash types.Hash
	copy(hash[:], data)
	data, err = bc.heights.Get(hash[:])
	if err != nil {
		return nil, errors.Wrap(err, "could not get height of current block")
	}
	block, err := bc.BlockByHash(hash)
	if err != nil {
		return nil, errors.Wrap(err, "could not get current block")
	}
	bc.height = binary.LittleEndian.Uint32(data)
	bc.block = block
	return bc, nil
}

// Height will return current highest height.
func (bc *Blockchain) Height() uint32 {
	return bc.height
}

// Header will return current highest header.
func (bc *Blockchain) Header() *types.Header {
	return bc.block.Header
}

// AddBlock adds a block to the blockchain.
func (bc *Blockchain) AddBlock(block *types.Block) error {

	// TODO: handle reorgs & best path block per height

	// check if we know the parent and get the parent height
	height, err := bc.HeightByHash(block.Parent)
	if err != nil {
		return errors.Wrap(err, "could not get parent header")
	}

	// calculate the hash once and prepare key for height lookups
	height++
	hash := block.Hash()
	data := make([]byte, 4)

	// mapp the block hash to the block height
	binary.LittleEndian.PutUint32(data, height)
	err = bc.heights.Put(hash[:], data)
	if err != nil {
		return errors.Wrap(err, "could not map block hash to height")
	}

	// map the block height to the block hash
	err = bc.heights.Put(data, hash[:])
	if err != nil {
		return errors.Wrap(err, "could not map block height to hash")
	}

	// save each transaction & collect their indices
	indices := make([]byte, 0, len(block.Transactions)*32)
	for _, tx := range block.Transactions {
		err = bc.transactions.Save(tx.Hash(), tx)
		if err != nil {
			return errors.Wrap(err, "could not store block transaction")
		}
		txHash := tx.Hash()
		indices = append(indices, txHash[:]...)
	}

	// map the block hash to the transaction indices
	err = bc.indices.Put(hash[:], indices)
	if err != nil {
		return errors.Wrap(err, "could not store block transaction IDs")
	}

	// save the block
	err = bc.headers.Save(block.Hash(), block.Header)
	if err != nil {
		return errors.Wrap(err, "could not store block header")
	}

	// if the block is not a new best block, we are done
	if height <= bc.height {
		return nil
	}

	// map the current height to the block hash
	binary.LittleEndian.PutUint32(data, Current)
	err = bc.heights.Put(data, hash[:])
	if err != nil {
		return errors.Wrap(err, "could not map current height to hash")
	}

	// store a reference to the block for fast access
	bc.height = height
	bc.block = block

	return nil
}

// HeightByHash returns the height of a block for the given hash.
func (bc *Blockchain) HeightByHash(hash types.Hash) (uint32, error) {
	data, err := bc.heights.Get(hash[:])
	if err != nil {
		return 0, errors.Wrap(err, "could not retrieve height by hash")
	}
	height := binary.LittleEndian.Uint32(data)
	return height, nil
}

// HashByHeight returns the hash of the block at the given height.
func (bc *Blockchain) HashByHeight(height uint32) (types.Hash, error) {
	key := make([]byte, 4)
	binary.LittleEndian.PutUint32(key, height)
	data, err := bc.heights.Get(key)
	if err != nil {
		return types.ZeroHash, errors.Wrap(err, "could not retrieve hash by height")
	}
	var hash types.Hash
	copy(hash[:], data)
	return hash, nil
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
func (bc *Blockchain) TransactionByHash(hash types.Hash) (*types.Transaction, error) {
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

// BlockByHeight retrieves a block by its height.
func (bc *Blockchain) BlockByHeight(height uint32) (*types.Block, error) {
	hash, err := bc.HashByHeight(height)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve hash for height")
	}
	return bc.BlockByHash(hash)
}

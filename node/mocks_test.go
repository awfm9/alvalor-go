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
	"io"

	"github.com/stretchr/testify/mock"

	"github.com/alvalor/alvalor-go/types"
)

type CodecMock struct {
	mock.Mock
}

func (c *CodecMock) Encode(w io.Writer, i interface{}) error {
	args := c.Called(w, i)
	return args.Error(0)
}

func (c *CodecMock) Decode(r io.Reader) (interface{}, error) {
	args := c.Called(r)
	return args.Get(0), args.Error(1)
}

type StoreMock struct {
	mock.Mock
}

func (s *StoreMock) Put(key []byte, data []byte) error {
	args := s.Called(key, data)
	return args.Error(0)
}

func (s *StoreMock) Get(key []byte) ([]byte, error) {
	args := s.Called(key)
	var data []byte
	if args.Get(0) != nil {
		data = args.Get(0).([]byte)
	}
	return data, args.Error(1)
}

func (s *StoreMock) Del(key []byte) error {
	args := s.Called(key)
	return args.Error(0)
}

type PoolMock struct {
	mock.Mock
}

func (p *PoolMock) Add(tx *types.Transaction) error {
	args := p.Called(tx)
	return args.Error(0)
}

func (p *PoolMock) Known(hash types.Hash) bool {
	args := p.Called(hash)
	return args.Bool(0)
}

func (p *PoolMock) Get(hash types.Hash) (*types.Transaction, error) {
	args := p.Called(hash)
	var tx *types.Transaction
	if args.Get(0) != nil {
		tx = args.Get(0).(*types.Transaction)
	}
	return tx, args.Error(1)
}

func (p *PoolMock) Remove(hash types.Hash) error {
	args := p.Called(hash)
	return args.Error(0)
}

func (p *PoolMock) Count() uint {
	args := p.Called()
	return uint(args.Int(0))
}

func (p *PoolMock) Hashes() []types.Hash {
	args := p.Called()
	var set []types.Hash
	if args.Get(0) != nil {
		set = args.Get(0).([]types.Hash)
	}
	return set
}

type PeersMock struct {
	mock.Mock
}

func (p *PeersMock) Active(address string) {
	p.Called(address)
}

func (p *PeersMock) Inactive(address string) {
	p.Called(address)
}

func (p *PeersMock) Actives() []string {
	args := p.Called()
	var active []string
	if args.Get(0) != nil {
		active = args.Get(0).([]string)
	}
	return active
}

func (p *PeersMock) Tag(address string, hash types.Hash) {
	p.Called(address, hash)
}

func (p *PeersMock) Tags(hash types.Hash) []string {
	args := p.Called(hash)
	var seen []string
	if args.Get(0) != nil {
		seen = args.Get(0).([]string)
	}
	return seen
}

type HandlersMock struct {
	mock.Mock
}

func (h *HandlersMock) Input(input <-chan interface{}) {
	h.Called(input)
}

func (h *HandlersMock) Event(event interface{}) {
	h.Called(event)
}

func (h *HandlersMock) Message(address string, message interface{}) {
	h.Called(address, message)
}

func (h *HandlersMock) Entity(entity Entity) {
	h.Called(entity)
}

func (h *HandlersMock) Collect(path []types.Hash) {
	h.Called(path)
}

type NetworkMock struct {
	mock.Mock
}

func (n *NetworkMock) Send(address string, msg interface{}) error {
	args := n.Called(address, msg)
	return args.Error(0)
}

func (n *NetworkMock) Broadcast(msg interface{}, exclude ...string) error {
	args := n.Called(msg, exclude)
	return args.Error(0)
}

type BlockchainMock struct {
	mock.Mock
}

func (b *BlockchainMock) Height() uint32 {
	args := b.Called()
	return uint32(args.Int(0))
}

func (b *BlockchainMock) Header() *types.Header {
	args := b.Called()
	return args.Get(0).(*types.Header)
}

func (b *BlockchainMock) AddBlock(block *types.Block) error {
	args := b.Called(block)
	return args.Error(0)
}

func (b *BlockchainMock) TransactionByHash(hash types.Hash) (*types.Transaction, error) {
	args := b.Called(hash)
	var tx *types.Transaction
	if args.Get(0) != nil {
		tx = args.Get(0).(*types.Transaction)
	}
	return tx, args.Error(1)
}

func (b *BlockchainMock) HeightByHash(hash types.Hash) (uint32, error) {
	args := b.Called(hash)
	return uint32(args.Int(0)), args.Error(1)
}

func (b *BlockchainMock) HashByHeight(height uint32) (types.Hash, error) {
	args := b.Called(height)
	return args.Get(0).(types.Hash), args.Error(1)
}

func (b *BlockchainMock) HeaderByHash(hash types.Hash) (*types.Header, error) {
	args := b.Called(hash)
	var header *types.Header
	if args.Get(0) != nil {
		header = args.Get(0).(*types.Header)
	}
	return header, args.Error(1)
}

func (b *BlockchainMock) HeaderByHeight(height uint32) (*types.Header, error) {
	args := b.Called(height)
	var header *types.Header
	if args.Get(0) != nil {
		header = args.Get(0).(*types.Header)
	}
	return header, args.Error(1)
}

func (b *BlockchainMock) BlockByHash(hash types.Hash) (*types.Block, error) {
	args := b.Called(hash)
	var block *types.Block
	if args.Get(0) != nil {
		block = args.Get(0).(*types.Block)
	}
	return block, args.Error(1)
}

func (b *BlockchainMock) BlockByHeight(height uint32) (*types.Block, error) {
	args := b.Called(height)
	var block *types.Block
	if args.Get(0) != nil {
		block = args.Get(0).(*types.Block)
	}
	return block, args.Error(1)
}

type FinderMock struct {
	mock.Mock
}

func (f *FinderMock) Add(hash types.Hash, parent types.Hash) error {
	args := f.Called(hash, parent)
	return args.Error(0)
}

func (f *FinderMock) Has(hash types.Hash) bool {
	args := f.Called(hash)
	return args.Bool(0)
}

func (f *FinderMock) Path() []types.Hash {
	args := f.Called()
	var path []types.Hash
	if args.Get(0) != nil {
		path = args.Get(0).([]types.Hash)
	}
	return path
}

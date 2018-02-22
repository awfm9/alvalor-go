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

	"github.com/alvalor/alvalor-go/types"
	"github.com/stretchr/testify/mock"
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

func (p *PoolMock) Known(id []byte) bool {
	args := p.Called(id)
	return args.Bool(0)
}

func (p *PoolMock) Get(id []byte) (*types.Transaction, error) {
	args := p.Called(id)
	var tx *types.Transaction
	if args.Get(0) != nil {
		tx = args.Get(0).(*types.Transaction)
	}
	return tx, args.Error(1)
}

func (p *PoolMock) Remove(id []byte) error {
	args := p.Called(id)
	return args.Error(0)
}

type StateMock struct {
	mock.Mock
}

func (s *StateMock) On(address string) {
	s.Called(address)
}

func (s *StateMock) Off(address string) {
	s.Called(address)
}

func (s *StateMock) Active() []string {
	args := s.Called()
	var active []string
	if args.Get(0) != nil {
		active = args.Get(0).([]string)
	}
	return active
}

func (s *StateMock) Tag(address string, id []byte) {
	s.Called(address, id)
}

func (s *StateMock) Seen(id []byte) []string {
	args := s.Called(id)
	var seen []string
	if args.Get(0) != nil {
		seen = args.Get(0).([]string)
	}
	return seen
}

type HandlersMock struct {
	mock.Mock
}

func (h *HandlersMock) Process(entity Entity) {
	h.Called(entity)
}

func (h *HandlersMock) Propagate(entity Entity) {
	h.Called(entity)
}

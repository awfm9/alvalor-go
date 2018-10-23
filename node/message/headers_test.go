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

package message

import (
	"github.com/alvalor/alvalor-go/types"
	"github.com/stretchr/testify/mock"
)

// HeadersMock mocks the header repository interface.
type HeadersMock struct {
	mock.Mock
}

// Get mocks the get function of the header repository interface.
func (hm *HeadersMock) Get(hash types.Hash) (*types.Header, error) {
	args := hm.Called(hash)
	var header *types.Header
	if args.Get(0) != nil {
		header = args.Get(0).(*types.Header)
	}
	return header, args.Error(1)
}

// Path mocks the path function of the header repository interface.
func (hm *HeadersMock) Path() ([]types.Hash, uint64) {
	args := hm.Called()
	var path []types.Hash
	if args.Get(0) != nil {
		path = args.Get(0).([]types.Hash)
	}
	return path, uint64(args.Int(1))
}

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

package event

import (
	"github.com/alvalor/alvalor-go/types"
	"github.com/stretchr/testify/mock"
)

// HeadersMock represents the mock for the headers store.
type HeadersMock struct {
	mock.Mock
}

// Path returns current best path with distance.
func (hm *HeadersMock) Path() ([]types.Hash, uint64) {
	args := hm.Called()
	var path []types.Hash
	if args.Get(0) != nil {
		path = args.Get(0).([]types.Hash)
	}
	return path, uint64(args.Int(1))
}

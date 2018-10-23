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

// TransactionsMock mocks the transactions repository interface.
type TransactionsMock struct {
	mock.Mock
}

// Has mocks the get function of the transactions repository interface.
func (tm *TransactionsMock) Get(hash types.Hash) (*types.Transaction, error) {
	args := tm.Called(hash)
	var tx *types.Transaction
	if args.Get(0) != nil {
		tx = args.Get(0).(*types.Transaction)
	}
	return tx, args.Error(1)
}

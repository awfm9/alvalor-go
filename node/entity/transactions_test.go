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

package entity

import (
	"github.com/alvalor/alvalor-go/types"
	"github.com/stretchr/testify/mock"
)

// TransactionsMock mocks the transactions repository interface.
type TransactionsMock struct {
	mock.Mock
}

// Add mocks the add function of the transactions repository interface.
func (tm *TransactionsMock) Add(transaction *types.Transaction) error {
	args := tm.Called(transaction)
	return args.Error(0)
}

// Has mocks the has function of the transactions repository interface.
func (tm *TransactionsMock) Has(hash types.Hash) bool {
	args := tm.Called(hash)
	return args.Bool(0)
}

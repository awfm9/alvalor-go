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

// EventMock mocks the events required by the entity package.
type EventMock struct {
	mock.Mock
}

// Header signals the reception of a new valid header.
func (em *EventMock) Header(header types.Hash) {
	em.Called(header)
}

// Transaction signals the reception of a new valid transaction.
func (em *EventMock) Transaction(transaction types.Hash) {
	em.Called(transaction)
}

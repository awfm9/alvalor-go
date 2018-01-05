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

package network

import (
	"github.com/stretchr/testify/mock"
)

type bookMock struct {
	mock.Mock
}

func (book *bookMock) Add(address string) {
	book.Called(address)
}
func (book *bookMock) Invalid(address string) {
	book.Called(address)
}
func (book *bookMock) Error(address string) {
	book.Called(address)
}
func (book *bookMock) Success(address string) {
	book.Called(address)
}
func (book *bookMock) Failure(address string) {
	book.Called(address)
}
func (book *bookMock) Dropped(address string) {
	book.Called(address)
}
func (book *bookMock) Sample(count int, filter func(*Entry) bool, less func(*Entry, *Entry) bool) ([]string, error) {
	args := book.Called(count, filter, less)
	return args.Get(0).([]string), args.Error(1)
}

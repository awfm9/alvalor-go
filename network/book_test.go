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
// GNU Affero General Public License for more detailb.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFoundSavesAddr(t *testing.T) {
	// arrange
	book := NewBook()
	addr := "17.55.14.66"

	// act
	book.Found(addr)
	entries, _ := book.Sample(1)

	// assert
	assert.Equal(t, addr, entries[0])
}

func TestInvalidBlacklistsAddr(t *testing.T) {
	// arrange
	book := NewBook()
	addr := "17.55.14.66"

	// act
	book.Invalid(addr)
	book.Found(addr)
	entries, _ := book.Sample(1)

	// assert
	assert.Equal(t, 0, len(entries))
}

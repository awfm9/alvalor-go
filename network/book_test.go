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
	"crypto/md5"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSampleReturnsErrorIfZeroCountPassed(t *testing.T) {
	// arrange
	book := NewBook()

	// act
	_, err := book.Sample(0)

	// assert
	assert.Equal(t, errInvalidCount, err)
}

func TestSampleReturnsErrorIfEmpty(t *testing.T) {
	// arrange
	book := NewBook()

	// act
	_, err := book.Sample(1)

	// assert
	assert.Equal(t, errBookEmpty, err)
}

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
	assert.Len(t, entries, 0)
}

func TestFailureDeactivatesAddress(t *testing.T) {
	// arrange
	book := NewBook()
	addr := "17.55.14.66"

	// act
	book.Found(addr)
	book.Failure(addr)
	entries, _ := book.Sample(1, isActive(true))

	// assert
	assert.Len(t, entries, 0)
}

func TestDroppedDeactivatesAddress(t *testing.T) {
	// arrange
	book := NewBook()
	addr := "17.55.14.66"

	// act
	book.Found(addr)
	book.Dropped(addr)
	entries, _ := book.Sample(1, isActive(true))

	// assert
	assert.Len(t, entries, 0)
}

func TestSampleReturnsAddressWithHighestScoreWhenOtherConnectionsDropped(t *testing.T) {
	// arrange
	book := NewBook()
	addr1 := "127.54.51.66"
	addr2 := "120.55.58.86"
	addr3 := "156.23.41.24"

	book.Found(addr1)
	book.Success(addr1)
	book.Error(addr1)
	book.Success(addr1)
	book.Error(addr1)

	book.Found(addr2)
	book.Success(addr2)
	book.Error(addr2)
	book.Success(addr2)

	book.Found(addr3)
	book.Success(addr3)
	book.Dropped(addr3)
	book.Success(addr3)

	entries, _ := book.Sample(10, isActive(true), byScore())

	assert.Len(t, entries, 3)
	assert.Equal(t, addr3, entries[0])
	assert.Equal(t, addr2, entries[1])
	assert.Equal(t, addr1, entries[2])
}

func TestSampleReturnsAddressWithHighestScoreWhenOtherConnectionsFailed(t *testing.T) {
	// arrange
	book := NewBook()
	addr1 := "127.54.51.66"
	addr2 := "120.55.58.86"
	addr3 := "156.23.41.24"

	book.Found(addr1)
	book.Success(addr1)
	book.Failure(addr1)
	book.Success(addr1)
	book.Failure(addr1)

	book.Found(addr2)
	book.Success(addr2)
	book.Failure(addr2)
	book.Success(addr2)

	book.Found(addr3)
	book.Success(addr3)

	entries, _ := book.Sample(10, isAny(), byScore())

	assert.Len(t, entries, 3)
	assert.Equal(t, addr3, entries[0])
	assert.Equal(t, addr2, entries[1])
	assert.Equal(t, addr1, entries[2])
}

func TestSampleReturnsAddressSortedRandomly(t *testing.T) {
	// arrange
	book := NewBook()
	addr1 := "127.54.51.66"
	addr2 := "120.55.58.86"
	addr3 := "156.23.41.24"
	addr4 := "177.7.44.62"

	book.Found(addr1)
	book.Found(addr2)
	book.Found(addr3)
	book.Found(addr4)

	entries1, _ := book.Sample(10, byRandom())
	entries2, _ := book.Sample(10, byRandom())

	assert.Len(t, entries1, 4)
	assert.Len(t, entries2, 4)
	equalCount := 0
	for i, value := range entries1 {
		if entries2[i] == value {
			equalCount++
		}
	}
	assert.NotEqual(t, 4, equalCount)
}

func TestSampleReturnsAddressSortedByHash(t *testing.T) {
	// arrange
	book := NewBook()
	addr1 := "127.54.51.66:27015"
	addr2 := "120.55.58.86:33523"

	book.Found(addr1)
	book.Found(addr2)

	entries, _ := book.Sample(10, byHash(func(data []byte) []byte {
		hasher := md5.New()
		hasher.Write(data)
		return hasher.Sum(nil)
	}))

	assert.Equal(t, addr1, entries[1])
	assert.Equal(t, addr2, entries[0])
}

func TestSampleReturnsOnlySpecifiedCountOfEntries(t *testing.T) {
	// arrange
	book := NewBook()
	addr1 := "127.54.51.66"
	addr2 := "120.55.58.86"
	addr3 := "156.23.41.24"

	book.Found(addr1)
	book.Found(addr2)
	book.Found(addr3)

	entries, _ := book.Sample(1)

	assert.Len(t, entries, 1)
}

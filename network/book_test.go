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
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddSavesAndGetsPeer(t *testing.T) {
	book := NewSimpleBook()
	address := "192.168.4.52"
	book.Add(address)

	entries, _ := book.Sample(1, IsActive(false), ByPrioritySort())

	assert.Equal(t, address, entries[0], "Entry %s was not found in book", address)
}

func TestOrderedSampleReturnsErr(t *testing.T) {
	book := NewSimpleBook()

	_, err := book.Sample(1, IsActive(false), ByPrioritySort())

	assert.NotNil(t, err, "Get should return error in case book is empty")
}

func TestAddDoesNotSavePeerIfBlacklisted(t *testing.T) {
	book := NewSimpleBook()
	address := "125.192.78.113"

	book.Invalid(address)
	book.Add(address)

	_, err := book.Sample(1, IsActive(false), ByPrioritySort())

	assert.NotNil(t, err, "Add should not save addressess in case it is blacklisted")
}

func TestOrderedSampleReturnsAddressWithHighestScoreWhenOtherConnectionsDropped(t *testing.T) {
	book := NewSimpleBook()
	address1 := "127.54.51.66"
	address2 := "120.55.58.86"
	address3 := "156.23.41.24"

	book.Add(address1)
	book.Add(address2)
	book.Add(address3)

	book.Success(address1)
	book.Dropped(address1)
	book.Success(address1)
	book.Dropped(address1)

	book.Success(address2)
	book.Dropped(address2)
	book.Success(address2)
	book.Dropped(address2)

	book.Success(address3)
	book.Dropped(address3)
	book.Success(address3)
	book.Dropped(address3)

	entries, _ := book.Sample(10, IsActive(false), ByPrioritySort())

	assert.Equal(t, address3, entries[0], "Address %s with highest score is expected. Actual addressess %s", address3, entries[0])
	assert.Equal(t, address2, entries[2])
	assert.Equal(t, address1, entries[1])
}

func TestOrderedSampleReturnsAddressWithHighestScoreWhenOtherConnectionsError(t *testing.T) {
	book := NewSimpleBook()
	address1 := "127.54.51.66"
	address2 := "120.55.58.86"
	address3 := "156.23.41.24"

	book.Add(address1)
	book.Add(address2)
	book.Add(address3)

	book.Error(address1)
	book.Error(address1)
	book.Error(address1)

	book.Error(address2)
	book.Error(address2)
	book.Error(address2)

	book.Success(address3)
	book.Dropped(address3)

	entries, _ := book.Sample(10, IsActive(false), ByPrioritySort())

	assert.Equal(t, address3, entries[0], "Address %s with highest score is expected. Actual addressess %s", address3, entries[0])
	assert.Equal(t, address2, entries[2])
	assert.Equal(t, address1, entries[1])
}

func TestOrderedSampleLimitsByParam(t *testing.T) {
	book := NewSimpleBook()
	address1 := "127.54.51.66"
	address2 := "120.55.58.86"
	address3 := "156.23.41.24"
	sampleSize := 2

	book.Add(address1)
	book.Add(address2)
	book.Add(address3)

	entries, _ := book.Sample(sampleSize, IsActive(false), ByPrioritySort())

	assert.Len(t, entries, sampleSize)
}

func TestRandomSampleReturnsErrorIfNoPeersAdded(t *testing.T) {
	book := NewSimpleBook()

	_, err := book.Sample(1, Any(), RandomSort())

	assert.NotNil(t, err)
}

func TestRandomSampleReturnsAddedPeers(t *testing.T) {
	book := NewSimpleBook()
	count := 10
	addresss := make([]string, 0, count)
	for i := 0; i < count; i++ {
		address := randomAddr()
		addresss = append(addresss, address)
		book.Add(address)
		book.Success(address)
	}

	sample, _ := book.Sample(count, Any(), RandomSort())

	assert.Subset(t, addresss, sample, "Expected sample to be a subset of addresss")
}

func TestRandomSampleReturnsSubsetOfAddedPeers(t *testing.T) {
	book := NewSimpleBook()
	count := 50
	addresss := make([]string, 0, count)
	for i := 0; i < count; i++ {
		address := randomAddr()
		addresss = append(addresss, address)
		book.Add(address)
		book.Success(address)
	}

	sample, _ := book.Sample(count, Any(), RandomSort())

	assert.Subset(t, addresss, sample, "Expected sample to be a subset of addresss")
}

func TestRandomSampleLimitsByParam(t *testing.T) {
	book := NewSimpleBook()
	address1 := "127.54.51.66"
	address2 := "120.55.58.86"
	address3 := "156.23.41.24"
	count := 2

	book.Add(address1)
	book.Add(address2)
	book.Add(address3)

	entries, _ := book.Sample(count, Any(), RandomSort())

	assert.Len(t, entries, count)
}

func randomAddr() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
}

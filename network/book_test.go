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
	addr := "192.168.4.52"
	book.Add(addr)

	entries, _ := book.Sample(1, IsActive(false), ByPrioritySort())

	assert.Equal(t, addr, entries[0], "Entry %s was not found in book", addr)
}

func TestOrderedSampleReturnsErr(t *testing.T) {
	book := NewSimpleBook()

	_, err := book.Sample(1, IsActive(false), ByPrioritySort())

	assert.NotNil(t, err, "Get should return error in case book is empty")
}

func TestAddDoesNotSavePeerIfBlacklisted(t *testing.T) {
	book := NewSimpleBook()
	addr := "125.192.78.113"

	book.Blacklist(addr)
	book.Add(addr)

	_, err := book.Sample(1, IsActive(false), ByPrioritySort())

	assert.NotNil(t, err, "Add should not save address in case it is blacklisted")
}

func TestAddSavesPeerIfBlacklistedAndWhitelistedLater(t *testing.T) {
	book := NewSimpleBook()
	addr := "125.192.78.113"

	book.Blacklist(addr)
	book.Whitelist(addr)
	book.Add(addr)

	entries, _ := book.Sample(1, IsActive(false), ByPrioritySort())

	assert.Equal(t, addr, entries[0], "Address should be saved after whitelisting")
}

func TestOrderedSampleReturnsAddressWithHighestScoreWhenOtherConnectionsDropped(t *testing.T) {
	book := NewSimpleBook()
	addr1 := "127.54.51.66"
	addr2 := "120.55.58.86"
	addr3 := "156.23.41.24"

	book.Add(addr1)
	book.Add(addr2)
	book.Add(addr3)

	book.Connected(addr1)
	book.Disconnected(addr1)
	book.Connected(addr1)
	book.Dropped(addr1)

	book.Connected(addr2)
	book.Dropped(addr2)
	book.Connected(addr2)
	book.Dropped(addr2)

	book.Connected(addr3)
	book.Disconnected(addr3)
	book.Connected(addr3)
	book.Disconnected(addr3)

	entries, _ := book.Sample(10, IsActive(false), ByPrioritySort())

	assert.Equal(t, addr3, entries[0], "Address %s with highest score is expected. Actual address %s", addr3, entries[0])
	assert.Equal(t, addr2, entries[2])
	assert.Equal(t, addr1, entries[1])
}

func TestOrderedSampleReturnsAddressWithHighestScoreWhenOtherConnectionsFailed(t *testing.T) {
	book := NewSimpleBook()
	addr1 := "127.54.51.66"
	addr2 := "120.55.58.86"
	addr3 := "156.23.41.24"

	book.Add(addr1)
	book.Add(addr2)
	book.Add(addr3)

	book.Failed(addr1)
	book.Failed(addr1)
	book.Failed(addr1)

	book.Failed(addr2)
	book.Failed(addr2)
	book.Failed(addr2)

	book.Connected(addr3)
	book.Disconnected(addr3)

	entries, _ := book.Sample(10, IsActive(false), ByPrioritySort())

	assert.Equal(t, addr3, entries[0], "Address %s with highest score is expected. Actual address %s", addr3, entries[0])
	assert.Equal(t, addr2, entries[2])
	assert.Equal(t, addr1, entries[1])
}

func TestOrderedSampleLimitsByParam(t *testing.T) {
	book := NewSimpleBook()
	addr1 := "127.54.51.66"
	addr2 := "120.55.58.86"
	addr3 := "156.23.41.24"
	sampleSize := 2

	book.Add(addr1)
	book.Add(addr2)
	book.Add(addr3)

	entries, _ := book.Sample(sampleSize, IsActive(false), ByPrioritySort())

	assert.Len(t, entries, sampleSize);
}

func TestRandomSampleReturnsErrorIfNoPeersAdded(t *testing.T) {
	book := NewSimpleBook()

	_, err := book.Sample(1, Any(), RandomSort())

	assert.NotNil(t, err)
}

func TestRandomSampleReturnsAddedPeers(t *testing.T) {
	book := NewSimpleBook()
	count := 10
	addrs := make([]string, 0, count)
	for i := 0; i < count; i++ {
		addr := randomAddr()
		addrs = append(addrs, addr)
		book.Add(addr)
		book.Connected(addr)
	}

	sample, _ := book.Sample(count, Any(), RandomSort())

	assert.Subset(t, addrs, sample, "Expected sample to be a subset of addrs")
}

func TestRandomSampleReturnsSubsetOfAddedPeers(t *testing.T) {
	book := NewSimpleBook()
	count := 50
	addrs := make([]string, 0, count)
	for i := 0; i < count; i++ {
		addr := randomAddr()
		addrs = append(addrs, addr)
		book.Add(addr)
		book.Connected(addr)
	}

	sample, _ := book.Sample(count, Any(), RandomSort())

	assert.Subset(t, addrs, sample, "Expected sample to be a subset of addrs")
}

func TestRandomSampleLimitsByParam(t *testing.T) {
	book := NewSimpleBook()
	addr1 := "127.54.51.66"
	addr2 := "120.55.58.86"
	addr3 := "156.23.41.24"
	count := 2

	book.Add(addr1)
	book.Add(addr2)
	book.Add(addr3)

	entries, _ := book.Sample(count, Any(), RandomSort())

	assert.Len(t, entries, count);
}

func randomAddr() string {
	return fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
}

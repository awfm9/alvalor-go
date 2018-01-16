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
	"bytes"
	"crypto/md5"
	"net"
	"reflect"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBook(t *testing.T) {

	book := NewBook()

	assert.NotNil(t, book.entries, "entries map not initialized")
	assert.NotNil(t, book.blacklist, "blacklist map not initialized")
}

func TestFound(t *testing.T) {

	book := NewBook()

	address := "17.55.14.66:7732"
	book.Found(address)
	_, known := book.entries[address]
	assert.True(t, known, "found didn't add valid address")

	address = "123.32.1.99:1373"
	book.blacklist[address] = struct{}{}
	book.Found(address)
	_, known = book.entries[address]
	assert.False(t, known, "found added blacklisted address")
}

func TestInvalid(t *testing.T) {

	book := NewBook()

	address := "17.55.14.66:7732"
	book.Invalid(address)
	_, blacklisted := book.blacklist[address]
	assert.True(t, blacklisted, "invalid didn't blacklist address")

	address = "123.32.1.99:1373"
	book.entries[address] = &entry{}
	book.Invalid(address)
	_, known := book.entries[address]
	assert.False(t, known, "invalid didn't remove address from entries")
}

func TestError(t *testing.T) {

	book := NewBook()

	address := "17.55.14.66:7732"
	book.Error(address)
	assert.Len(t, book.entries, 0, "error on unknown address modified state")

	e := &entry{Failure: 0}
	book.entries[address] = e
	book.Error(address)
	assert.Equal(t, e.Failure, 1, "error did not increase failure counter")
}

func TestSuccess(t *testing.T) {

	book := NewBook()

	address := "17.55.14.66:7732"
	book.Success(address)
	assert.Len(t, book.entries, 0, "success on unknown address modified state")

	e := &entry{Success: 0, Active: false}
	book.entries[address] = e
	book.Success(address)
	assert.Equal(t, e.Success, 1, "success did not increase success counter")
	assert.True(t, e.Active, "success did not change entry to active")
}

func TestDropped(t *testing.T) {

	book := NewBook()

	address := "17.55.14.66:7732"
	book.Dropped(address)
	assert.Len(t, book.entries, 0, "dropped on unknown address modified state")

	e := &entry{Active: true}
	book.entries[address] = e
	book.Dropped(address)
	assert.False(t, e.Active, "dropped did not change entry to inactive")
}

func TestFailure(t *testing.T) {

	book := NewBook()

	address := "17.55.14.66:7732"
	book.Failure(address)
	assert.Len(t, book.entries, 0, "failure on unknown address modified state")

	e := &entry{Failure: 0, Active: true}
	book.entries[address] = e
	book.Failure(address)
	assert.Equal(t, e.Failure, 1, "failure did not increase failure counter")
	assert.False(t, e.Active, "failure did not change entry to inactive")
}

func TestSample(t *testing.T) {

	book := NewBook()
	addr1 := "192.0.2.1:1337"
	addr2 := "192.0.2.2:1337"
	addr3 := "192.0.2.3:1337"
	addr4 := "192.0.2.4:1337"
	addr5 := "192.0.2.5:1337"
	addr6 := "192.0.2.6:1337"
	addr7 := "192.0.2.7:1337"

	book.entries[addr1] = &entry{Address: addr1, Success: 1, Failure: 0, Active: true}  // +1
	book.entries[addr2] = &entry{Address: addr2, Success: 0, Failure: 7, Active: false} // -7
	book.entries[addr3] = &entry{Address: addr3, Success: 2, Failure: 5, Active: true}  // -3
	book.entries[addr4] = &entry{Address: addr4, Success: 0, Failure: 0, Active: false} // +0
	book.entries[addr5] = &entry{Address: addr5, Success: 5, Failure: 2, Active: true}  // +3
	book.entries[addr6] = &entry{Address: addr6, Success: 5, Failure: 0, Active: false} // +5
	book.entries[addr7] = &entry{Address: addr7, Success: 0, Failure: 1, Active: true}  // -1

	actual := book.Sample(6)
	assert.Len(t, actual, 6, "undersampling returns invalid count")

	actual = book.Sample(7)
	assert.Len(t, actual, 7, "exact sampling returns invalid count")

	actual = book.Sample(8)
	assert.Len(t, actual, 7, "oversampling returns invalid count")

	actual = book.Sample(7, isAny())
	expected := []string{addr1, addr2, addr3, addr4, addr5, addr6, addr7}
	assert.ElementsMatch(t, expected, actual, "is any filter returns wrong elements")

	actual = book.Sample(7, isActive(true))
	expected = []string{addr1, addr3, addr5, addr7}
	assert.ElementsMatch(t, expected, actual, "is active filter returns wrong elements")

	actual = book.Sample(7, byScore())
	expected = []string{addr6, addr5, addr1, addr4, addr7, addr3, addr2}
	assert.Equal(t, expected, actual, "by score sort returns wrong ordering")

	actual = book.Sample(7, byHash(func(data []byte) []byte {
		hasher := md5.New()
		hasher.Write(data)
		return hasher.Sum(nil)
	}))
	expected = []string{addr1, addr2, addr3, addr4, addr5, addr6, addr7}
	sort.Slice(expected, func(i int, j int) bool {
		ip1, _, _ := net.SplitHostPort(expected[i])
		ip2, _, _ := net.SplitHostPort(expected[j])
		h1 := md5.Sum([]byte(ip1))
		h2 := md5.Sum([]byte(ip2))
		return bytes.Compare(h1[:], h2[:]) < 0
	})
	assert.Equal(t, expected, actual, "by hash sort has wrong order")

	mismatch := false
	for i := 0; i < 100; i++ {
		sample1 := book.Sample(7, byRandom())
		sample2 := book.Sample(7, byRandom())
		if !reflect.DeepEqual(sample1, sample2) {
			mismatch = true
			break
		}
	}
	assert.True(t, mismatch, "by random sorts all equal")
}

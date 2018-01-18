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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBook(t *testing.T) {

	book := NewBook()

	assert.NotNil(t, book.entries, "Entries map not initialized")
	assert.NotNil(t, book.blacklist, "Blacklist map not initialized")
}

func TestFound(t *testing.T) {

	book := NewBook()

	address := "192.0.2.1:1337"
	book.Found(address)
	_, known := book.entries[address]
	assert.True(t, known, "Found didn't add valid address to entries")

	address = "192.0.2.2:1337"
	book.blacklist[address] = struct{}{}
	book.Found(address)
	_, known = book.entries[address]
	assert.False(t, known, "Found added blacklisted address to entries")
}

func TestInvalid(t *testing.T) {

	book := NewBook()

	address := "192.0.2.1:1337"
	book.Invalid(address)
	_, blacklisted := book.blacklist[address]
	assert.True(t, blacklisted, "Invalid didn't blacklist address")

	address = "192.0.2.2:1337"
	book.entries[address] = &entry{}
	book.Invalid(address)
	_, known := book.entries[address]
	assert.False(t, known, "Invalid didn't remove address from entries")
}

func TestError(t *testing.T) {

	book := NewBook()

	address := "192.0.2.1:1337"
	book.Error(address)
	assert.Len(t, book.entries, 0, "Error on unknown address modified entries")

	e := &entry{Failure: 0}
	book.entries[address] = e
	book.Error(address)
	assert.Equal(t, e.Failure, 1, "Error did not increase failure counter on entry")
}

func TestSuccess(t *testing.T) {

	book := NewBook()

	address := "192.0.2.1:1337"
	book.Success(address)
	assert.Len(t, book.entries, 0, "Success on unknown address modified entries")

	e := &entry{Success: 0, Active: false}
	book.entries[address] = e
	book.Success(address)
	assert.Equal(t, e.Success, 1, "Success did not increase success counter on entry")
	assert.True(t, e.Active, "Success did not change entry to active")
}

func TestDropped(t *testing.T) {

	book := NewBook()

	address := "192.0.2.1:1337"
	book.Dropped(address)
	assert.Len(t, book.entries, 0, "Dropped on unknown address modified entries")

	e := &entry{Active: true}
	book.entries[address] = e
	book.Dropped(address)
	assert.False(t, e.Active, "Dropped did not change entry to inactive")
}

func TestFailure(t *testing.T) {

	book := NewBook()

	address := "192.0.2.1:1337"
	book.Failure(address)
	assert.Len(t, book.entries, 0, "Failure on unknown address modified entries")

	e := &entry{Failure: 0, Active: true}
	book.entries[address] = e
	book.Failure(address)
	assert.Equal(t, e.Failure, 1, "Failure did not increase failure counter on entry")
	assert.False(t, e.Active, "Failure did not change entry to inactive")
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

	book.entries[addr1] = &entry{Address: addr1, Success: 1, Failure: 0, Active: true}
	book.entries[addr2] = &entry{Address: addr2, Success: 0, Failure: 7, Active: false}
	book.entries[addr3] = &entry{Address: addr3, Success: 2, Failure: 5, Active: true}
	book.entries[addr4] = &entry{Address: addr4, Success: 0, Failure: 0, Active: false}
	book.entries[addr5] = &entry{Address: addr5, Success: 5, Failure: 2, Active: true}
	book.entries[addr6] = &entry{Address: addr6, Success: 5, Failure: 0, Active: false}
	book.entries[addr7] = &entry{Address: addr7, Success: 0, Failure: 1, Active: true}

	actual := book.Sample(6)
	assert.Len(t, actual, 6, "Undersampling returns invalid count")

	actual = book.Sample(7)
	assert.Len(t, actual, 7, "Exact sampling returns invalid count")

	actual = book.Sample(8)
	assert.Len(t, actual, 7, "Oversampling returns invalid count")

	actual = book.Sample(7, isAny())
	expected := []string{addr1, addr2, addr3, addr4, addr5, addr6, addr7}
	assert.ElementsMatch(t, expected, actual, "Is any filter returns wrong elements")

	actual = book.Sample(7, isActive(true))
	expected = []string{addr1, addr3, addr5, addr7}
	assert.ElementsMatch(t, expected, actual, "Is active filter returns wrong elements")

	actual = book.Sample(7, byScore())
	expected = []string{
		addr6, // +5
		addr5, // +3
		addr1, // +1
		addr4, // +0
		addr7, // -1
		addr3, // -3
		addr2, // -7
	}
	assert.Equal(t, expected, actual, "By score sort returns wrong ordering")

	actual = book.Sample(7, byHashFunc(md5.New()))
	expected = []string{
		addr2, // 7f83fddecaba901abfd469d899958433
		addr5, // 87cf214a6664ce35b0b247b4e8c4a529
		addr6, // d07d5578976452b56c7be96b384603cd
		addr1, // d0f88d6c87767262ba8e93d6acccd784
		addr7, // e9cc93d8ae23cfc5941cec7cc4ee788a
		addr4, // ecf7a94a85ef4966739abcf2f32668bf
		addr3, // f2d19c83feb2cb2458fd427acfb343d7
	}
	assert.Equal(t, expected, actual, "By hash sort returns wrong ordering")

	mismatch := false
	for i := 0; i < 100; i++ {
		sample1 := book.Sample(7, byRandom())
		sample2 := book.Sample(7, byRandom())
		if !reflect.DeepEqual(sample1, sample2) {
			mismatch = true
			break
		}
	}
	assert.True(t, mismatch, "By random sort always returns same ordering")
}

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
	"testing"
)

func TestAddSavesPeer(t *testing.T) {
	book := DefaultBook
	addr := "192.168.4.52"
	book.Add(addr)

	entry, _ := book.Get()

	if entry != addr {
		t.Fatalf("Entry %s was not found in book", addr)
	}
}

func TestGetReturnsErrBookEmpty(t *testing.T) {
	book := DefaultBook

	_, err := book.Get()

	if err != errBookEmpty {
		t.Fatalf("Get should return error in case book is empty")
	}
}

func TestAddDoesNotSavePeerIfBlacklisted(t *testing.T) {
	book := DefaultBook
	addr := "125.192.78.113"

	book.Blacklist(addr)
	book.Add(addr)

	_, err := book.Get()

	if err != errBookEmpty {
		t.Fatalf("Add should not save address in case it is blacklisted")
	}
}

func TestAddSavesPeerIfBlacklistedAndWhitelistedLater(t *testing.T) {
	book := DefaultBook
	addr := "125.192.78.113"

	book.Blacklist(addr)
	book.Whitelist(addr)
	book.Add(addr)

	entry, _ := book.Get()

	if entry != addr {
		t.Fatalf("Address should be saved after whitelisting")
	}
}

func TestGetReturnsAddressWithHighestScore(t *testing.T) {
	book := DefaultBook
	addr1 := "127.54.51.66"
	addr2 := "120.55.58.86"
	addr3 := "156.23.41.24"

	book.Add(addr1)
    book.Connected(addr1)
	book.Dropped(addr1)

	book.Add(addr2)
    book.Connected(addr2)
	book.Dropped(addr2)
	book.Connected(addr2)
	book.Dropped(addr2)

	book.Add(addr3)
	book.Connected(addr3)
	book.Disconnected(addr3)
	book.Connected(addr3)
	book.Disconnected(addr3)
	book.Connected(addr3)

	entry, _ := book.Get()

	if entry != addr3 {
		t.Fatalf("Address %s with highest score is expected. Actual address %s", addr3, entry)
	}
}
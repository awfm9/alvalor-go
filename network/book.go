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
	"bytes"
	"math/rand"
	"net"
)

// Book represents an address book interface to handle known peer addresses on the alvalor
type Book interface {
	Add(address string)
	Invalid(address string)
	Error(address string)
	Success(address string)
	Failure(address string)
	Dropped(address string)
	Sample(count int, filter func(*Entry) bool, less func(*Entry, *Entry) bool) ([]string, error)
}

// Entry represents an entry in the simple address book, containing the address, whether the peer is
// currently active and how many successes/failures we had on the connection.
type Entry struct {
	Address string
	Active  bool
	Success int
	Failure int
}

// IsActive represents filter to select active/inactive entries in Sample method
func IsActive(active bool) func(e *Entry) bool {
	return func(e *Entry) bool {
		return e.Active == active
	}
}

// IsAny reperesents filter to select any entries in Sample method
func IsAny() func(e *Entry) bool {
	return func(e *Entry) bool {
		return true
	}
}

// ByScore represents an order by priority which is calculated based on score. It can be used to sort entries in Sample method
func ByScore(score func(*Entry) float64) func(*Entry, *Entry) bool {
	return func(e1 *Entry, e2 *Entry) bool {
		return score(e1) > score(e2)
	}
}

// ByRandom represents a random order. It can be used to sort entries in Sample method
func ByRandom() func(*Entry, *Entry) bool {
	return func(*Entry, *Entry) bool {
		return rand.Int()%2 == 0
	}
}

// ByHash represents ordering by IP hash, to distribute geographically.
func ByHash(hash func([]byte) []byte) func(*Entry, *Entry) bool {
	return func(e1 *Entry, e2 *Entry) bool {
		ip1, _, _ := net.SplitHostPort(e1.Address)
		ip2, _, _ := net.SplitHostPort(e2.Address)
		h1 := hash([]byte(ip1))
		h2 := hash([]byte(ip2))
		return bytes.Compare(h1[:], h2[:]) < 0
	}
}

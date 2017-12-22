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

package sort

import (
	"bytes"
	"math/rand"
	"net"

	"github.com/alvalor/alvalor-go/book"
)

// Score represents an order by priority which is calculated based on score. It can be used to sort entries in Sample method
func Score(score func(*book.Entry) float64) func(*book.Entry, *book.Entry) bool {
	return func(e1 *book.Entry, e2 *book.Entry) bool {
		return score(e1) > score(e2)
	}
}

// Random represents a random order. It can be used to sort entries in Sample method
func Random() func(*book.Entry, *book.Entry) bool {
	return func(*book.Entry, *book.Entry) bool {
		return rand.Int()%2 == 0
	}
}

// Hash represents ordering by IP hash, to distribute geographically.
func Hash(hash func([]byte) []byte) func(*book.Entry, *book.Entry) bool {
	return func(e1 *book.Entry, e2 *book.Entry) bool {
		ip1, _, _ := net.SplitHostPort(e1.Address)
		ip2, _, _ := net.SplitHostPort(e2.Address)
		h1 := hash([]byte(ip1))
		h2 := hash([]byte(ip2))
		return bytes.Compare(h1[:], h2[:]) < 0
	}
}

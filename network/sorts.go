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
	"math"
	"math/rand"
	"net"
)

type sortFunc func(*entry, *entry) bool

func byScoreFunc(score func(*entry) float64) func(*entry, *entry) bool {
	return func(e1 *entry, e2 *entry) bool {
		return score(e1) > score(e2)
	}
}

func byScore() func(*entry, *entry) bool {
	return byScoreFunc(func(entry *entry) float64 {
		return math.Min((float64(entry.Success)/float64(entry.Failure))/100, 1)
	})
}

func byRandom() func(*entry, *entry) bool {
	return func(*entry, *entry) bool {
		return rand.Int()%2 == 0
	}
}

func byHash(hash func([]byte) []byte) func(*entry, *entry) bool {
	return func(e1 *entry, e2 *entry) bool {
		ip1, _, _ := net.SplitHostPort(e1.Address)
		ip2, _, _ := net.SplitHostPort(e2.Address)
		h1 := hash([]byte(ip1))
		h2 := hash([]byte(ip2))
		return bytes.Compare(h1[:], h2[:]) < 0
	}
}

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
	"hash"
	"math/rand"
	"net"
)

func byRandom() func(string, string) bool {
	return func(string, string) bool {
		return rand.Int()%2 == 0
	}
}

func byScore(rep reputationManager) func(string, string) bool {
	return func(address1 string, address2 string) bool {
		return rep.Score(address1) > rep.Score(address2)
	}
}

func byIPHash(h hash.Hash) func(string, string) bool {
	return func(address1 string, address2 string) bool {
		host1, _, _ := net.SplitHostPort(address1)
		h.Reset()
		_, _ = h.Write([]byte(host1))
		hash1 := h.Sum(nil)
		host2, _, _ := net.SplitHostPort(address2)
		h.Reset()
		_, _ = h.Write([]byte(host2))
		hash2 := h.Sum(nil)
		return bytes.Compare(hash1[:], hash2[:]) < 0
	}
}

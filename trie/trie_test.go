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

package trie

import (
	"bytes"
	"math/rand"
	"testing"
	"time"
)

func TestAll(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	trie := New()
	for i := 0; i < 1000; i++ {
		key := make([]byte, 32)
		hash := make([]byte, 32)
		_, _ = rand.Read(key)
		_, _ = rand.Read(hash)
		ok := trie.Put(key, hash, false)
		if !ok {
			t.Fatalf("could not put: %x", key)
		}
		out, ok := trie.Get(key)
		if !ok {
			t.Fatalf("could not get: %x", key)
		}
		if !bytes.Equal(out, hash) {
			t.Errorf("wrong hash: %x != %x", out, hash)
		}
		ok = trie.Del(key)
		if !ok {
			t.Fatalf("could not del: %x", key)
		}
		_, ok = trie.Get(key)
		if ok {
			t.Fatalf("should not get: %x", key)
		}
	}
}

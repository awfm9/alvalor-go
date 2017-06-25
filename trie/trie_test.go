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

	"github.com/alvalor/alvalor-go/hasher"
)

func TestSingle(t *testing.T) {
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

func TestBatch(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	trie := New()
	keys := make([][]byte, 0, 1000)
	hashes := make([][]byte, 0, 1000)
	for i := 0; i < 1000; i++ {
		key := make([]byte, 32)
		hash := make([]byte, 32)
		_, _ = rand.Read(key)
		_, _ = rand.Read(hash)
		keys = append(keys, key)
		hashes = append(hashes, hash)
	}
	for i, key := range keys {
		hash := hashes[i]
		ok := trie.Put(key, hash, false)
		if !ok {
			t.Fatalf("could not put %v: %x", i, key)
		}
	}
	for i, key := range keys {
		out, ok := trie.Get(key)
		if !ok {
			t.Fatalf("could not get %v: %x", i, key)
		}
		hash := hashes[i]
		if !bytes.Equal(out, hash) {
			t.Errorf("wrong hash: %x != %x", out, hash)
		}
	}
	for i, key := range keys {
		ok := trie.Del(key)
		if !ok {
			t.Fatalf("could not del %v: %x", i, key)
		}
	}
	if !bytes.Equal(trie.Hash(), hasher.Zero256) {
		t.Fatalf("root hash not zero: %x", trie.Hash())
	}
}

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

const HexTestLength = 100000

func TestHexSingle(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	trie := NewHex()
	for i := 0; i < HexTestLength; i++ {
		key := make([]byte, 32)
		hash := make([]byte, 32)
		_, _ = rand.Read(key)
		_, _ = rand.Read(hash)
		err := trie.Put(key, hash)
		if err != nil {
			t.Fatalf("could not put: %x", key)
		}
		out, err := trie.Get(key)
		if err != nil {
			t.Fatalf("could not get: %x", key)
		}
		if !bytes.Equal(out, hash) {
			t.Errorf("wrong hash: %x != %x", out, hash)
		}
		err = trie.Del(key)
		if err != nil {
			t.Fatalf("could not del: %x", key)
		}
		_, err = trie.Get(key)
		if err == nil {
			t.Fatalf("should not get: %x", key)
		}
	}
}

func TestHexBatch(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	trie := NewHex()
	keys := make([][]byte, 0, HexTestLength)
	hashes := make([][]byte, 0, HexTestLength)
	for i := 0; i < HexTestLength; i++ {
		key := make([]byte, 32)
		hash := make([]byte, 32)
		_, _ = rand.Read(key)
		_, _ = rand.Read(hash)
		keys = append(keys, key)
		hashes = append(hashes, hash)
	}
	for i, key := range keys {
		hash := hashes[i]
		err := trie.Put(key, hash)
		if err != nil {
			t.Fatalf("could not put %v: %x", i, key)
		}
	}
	for i, key := range keys {
		out, err := trie.Get(key)
		if err != nil {
			t.Fatalf("could not get %v: %x", i, key)
		}
		hash := hashes[i]
		if !bytes.Equal(out, hash) {
			t.Errorf("wrong hash: %x != %x", out, hash)
		}
	}
	for i, key := range keys {
		err := trie.Del(key)
		if err != nil {
			t.Fatalf("could not del %v: %x", i, key)
		}
	}
	zero := trie.h.Sum(nil)
	hash := trie.Hash()
	if !bytes.Equal(hash, zero) {
		t.Fatalf("root hash not zero: %x != %x", hash, zero)
	}
}

func BenchmarkHexInsert(b *testing.B) {
	t := NewHex()
	keys := make([][]byte, 0, b.N)
	hashes := make([][]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, 32)
		hash := make([]byte, 32)
		_, _ = rand.Read(key)
		_, _ = rand.Read(hash)
		keys = append(keys, key)
		hashes = append(hashes, hash)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.MustPut(keys[i], hashes[i])
	}
}

func BenchmarkHexRetrieve(b *testing.B) {
	t := NewHex()
	keys := make([][]byte, 0, b.N)
	hashes := make([][]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, 32)
		hash := make([]byte, 32)
		_, _ = rand.Read(key)
		_, _ = rand.Read(hash)
		keys = append(keys, key)
		hashes = append(hashes, hash)
	}
	for i := 0; i < b.N; i++ {
		t.MustPut(keys[i], hashes[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Get(keys[i])
	}
}

func BenchmarkHexDelete(b *testing.B) {
	t := NewHex()
	keys := make([][]byte, 0, b.N)
	hashes := make([][]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, 32)
		hash := make([]byte, 32)
		_, _ = rand.Read(key)
		_, _ = rand.Read(hash)
		keys = append(keys, key)
		hashes = append(hashes, hash)
	}
	for i := 0; i < b.N; i++ {
		t.MustPut(keys[i], hashes[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Del(keys[i])
	}
}

func BenchmarkHexHash(b *testing.B) {
	t := NewHex()
	keys := make([][]byte, 0, b.N)
	hashes := make([][]byte, 0, b.N)
	for i := 0; i < b.N; i++ {
		key := make([]byte, 32)
		hash := make([]byte, 32)
		_, _ = rand.Read(key)
		_, _ = rand.Read(hash)
		keys = append(keys, key)
		hashes = append(hashes, hash)
	}
	for i := 0; i < b.N; i++ {
		t.MustPut(keys[i], hashes[i])
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		t.Hash()
	}
}

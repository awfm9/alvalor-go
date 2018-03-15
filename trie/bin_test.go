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

const BinTestLength = 100000

func TestBinSingle(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	trie := NewBin()
	zero := trie.h.Sum(nil)
	for i := 0; i < BinTestLength; i++ {
		key := make([]byte, 32)
		hash := make([]byte, 32)
		_, _ = rand.Read(key)
		_, _ = rand.Read(hash)
		err := trie.Put(key, hash)
		if err != nil {
			t.Fatalf("could not put: %x (%v)", key, err)
		}
		out, err := trie.Get(key)
		if err != nil {
			t.Fatalf("could not get: %x (%v)", key, err)
		}
		if !bytes.Equal(out, hash) {
			t.Errorf("wrong hash: %x != %x", out, hash)
		}
		err = trie.Del(key)
		if err != nil {
			t.Fatalf("could not del: %x (%v)", key, err)
		}
		_, err = trie.Get(key)
		if err == nil {
			t.Fatalf("should not get: %x (%v)", key, err)
		}
		hash = trie.Hash()
		if !bytes.Equal(hash, zero) {
			t.Fatalf("root hash not zero: %x != %x", hash, zero)
		}
	}
}

func TestBinBatch(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	trie := NewBin()
	keys := make([][]byte, 0, BinTestLength)
	hashes := make([][]byte, 0, BinTestLength)
	for i := 0; i < BinTestLength; i++ {
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
			t.Fatalf("could not put %v: %x (%v)", i, key, err)
		}
	}
	for i, key := range keys {
		out, err := trie.Get(key)
		if err != nil {
			t.Fatalf("could not get %v: %x (%v)", i, key, err)
		}
		hash := hashes[i]
		if !bytes.Equal(out, hash) {
			t.Errorf("wrong hash: %x != %x", out, hash)
		}
	}
	for i, key := range keys {
		err := trie.Del(key)
		if err != nil {
			t.Fatalf("could not del %v: %x (%v)", i, key, err)
		}
	}
	zero := trie.h.Sum(nil)
	hash := trie.Hash()
	if !bytes.Equal(hash, zero) {
		t.Fatalf("root hash not zero: %x != %x", hash, zero)
	}
}

func BenchmarkBinInsert(b *testing.B) {
	t := NewBin()
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

func BenchmarkBinRetrieve(b *testing.B) {
	t := NewBin()
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

func BenchmarkBinDelete(b *testing.B) {
	t := NewBin()
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

func BenchmarkBinHash(b *testing.B) {
	t := NewBin()
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

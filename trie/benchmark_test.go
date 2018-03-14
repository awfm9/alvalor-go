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
	"crypto/rand"
	"testing"
)

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

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

package types

import (
	"encoding/binary"
	"time"

	"golang.org/x/crypto/blake2s"
)

// Header represents the header data of a block that will be hashed.
type Header struct {
	Parent Hash
	State  Hash
	Delta  Hash
	Miner  Hash
	Diff   uint64
	Nonce  uint64
	Time   time.Time
	Hash   Hash
}

func (hdr Header) calc() {
	h, _ := blake2s.New256(nil)
	_, _ = h.Write(hdr.Parent[:])
	_, _ = h.Write(hdr.State[:])
	_, _ = h.Write(hdr.Delta[:])
	_, _ = h.Write(hdr.Miner[:])
	data := make([]byte, 24)
	binary.LittleEndian.PutUint64(data[:8], hdr.Diff)
	binary.LittleEndian.PutUint64(data[8:16], uint64(hdr.Time.Unix()))
	binary.LittleEndian.PutUint64(data[16:], hdr.Nonce)
	_, _ = h.Write(data)
	hash := h.Sum(nil)
	copy(hdr.Hash[:], hash)
}

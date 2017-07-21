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

package transaction

import (
	"encoding/binary"

	"golang.org/x/crypto/blake2b"
)

// Fee represents a signed fee transaction.
type Fee struct {
	Origin []byte // account issuing the fee transaction
	Target []byte // account or action ID that the fee is meant for
	Amount uint64 // amount made available in fees
	Nonce  uint64 // nonce to doubling fee on same target
}

// ID returns the ID of the fee transaction.
func (f Fee) ID() []byte {
	h, _ := blake2b.New256(nil)
	h.Write(f.Origin)
	h.Write(f.Target)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, f.Amount)
	h.Write(buf)
	binary.LittleEndian.PutUint64(buf, f.Nonce)
	h.Write(buf)
	return h.Sum(nil)
}

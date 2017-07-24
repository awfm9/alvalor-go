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

// Transfer represents a value transfer transaction.
type Transfer struct {
	From   []byte // origin account of the transfer
	To     []byte // target account of the transfer
	Amount uint64 // amount to be transfered
	Nonce  uint64 // nonce to make sure we can issue same transfer again
}

// ID returns the ID of the transfer.
func (t Transfer) ID() []byte {
	h, _ := blake2b.New256(nil)
	h.Write(t.From)
	h.Write(t.To)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, t.Amount)
	h.Write(buf)
	binary.LittleEndian.PutUint64(buf, t.Nonce)
	h.Write(buf)
	return h.Sum(nil)
}

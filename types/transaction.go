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

	"golang.org/x/crypto/blake2b"
)

// Transaction represents an atomic standard transaction on the Alvalor network.
type Transaction struct {
	Transfers  []*Transfer
	Fees       []*Fee
	Data       []byte
	Signatures [][]byte
}

// Hash returns the unique hash of the transaction.
func (tx Transaction) Hash() []byte {
	buf := make([]byte, 8)
	h, _ := blake2b.New256(nil)
	for _, transfer := range tx.Transfers {
		_, _ = h.Write(transfer.From)
		_, _ = h.Write(transfer.To)
		binary.LittleEndian.PutUint64(buf, transfer.Amount)
		_, _ = h.Write(buf)
	}
	for _, fee := range tx.Fees {
		_, _ = h.Write(fee.From)
		binary.LittleEndian.PutUint64(buf, fee.Amount)
		_, _ = h.Write(buf)
	}
	_, _ = h.Write(tx.Data)
	return h.Sum(nil)
}

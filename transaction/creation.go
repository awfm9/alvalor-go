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

// Creation creates a new account.
type Creation struct {
	Origin      []byte   // account issuing the creation of a new account
	Target      []byte   // target account to be created
	Required    uint64   // number of signatories required for a valid signature
	Signatories [][]byte // public keys of the valid account signatories
}

// ID returns the ID of the cration transaction.
func (c Creation) ID() []byte {
	h, _ := blake2b.New256(nil)
	h.Write(c.Origin)
	h.Write(c.Target)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, c.Required)
	h.Write(buf)
	for _, signatory := range c.Signatories {
		h.Write(signatory)
	}
	return h.Sum(nil)
}

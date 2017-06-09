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

package hasher

import "golang.org/x/crypto/blake2b"

// Zero256 represents the 256-bit hash of an empty byte array.
var Zero256 = Sum256(nil)

// Zero512 represents the 512-bit hash of an empty byte array.
var Zero512 = Sum512(nil)

// Sum256 returns the 256-bit blake2b hash of the input data.
func Sum256(data []byte) []byte {
	h, _ := blake2b.New256(nil)
	h.Write(data)
	hash := h.Sum(nil)
	return hash
}

// Sum512 returns the 512-bit blake2b hash of the input data.
func Sum512(data []byte) []byte {
	h, _ := blake2b.New512(nil)
	h.Write(data)
	hash := h.Sum(nil)
	return hash
}

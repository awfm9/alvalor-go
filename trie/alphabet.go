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

// encode takes a byte slice and turns it into a slice with twice the length, from bytes with 256
// possible values each, to twice the number of nibbles with 16 possible values each. This can then
// be used to traverse down our patricia merkle trie with 16 children per node.
func encode(key []byte) []byte {
	path := make([]byte, 0, len(key)*2)
	for _, b := range key {
		path = append(path, b/16)
		path = append(path, b%16)
	}
	return path
}

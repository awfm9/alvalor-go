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

package subscribe

import "github.com/alvalor/alvalor-go/types"

type msgType uint8

const (
	hdrType msgType = iota
	txType
)

func isType(types ...msgType) func(interface{}) bool {
	return func(msg interface{}) bool {
		switch msg.(type) {
		case *Header:
			return containsType(types, hdrType)
		case *Transaction:
			return containsType(types, txType)
		}
		return false
	}
}

func containsType(types []msgType, inputType msgType) bool {
	for _, curType := range types {
		if inputType == curType {
			return true
		}
	}
	return false
}

func hasHash(hashes ...types.Hash) func(interface{}) bool {
	return func(msg interface{}) bool {
		tx, ok := msg.(*Transaction)
		if ok && containsHash(hashes, tx.hash) {
			return true
		}
		hdr, ok := msg.(*Header)
		if ok && containsHash(hashes, hdr.hash) {
			return true
		}
		return false
	}
}

func containsHash(hashes []types.Hash, hash types.Hash) bool {
	for _, hsh := range hashes {
		if hsh == hash {
			return true
		}
	}
	return false
}

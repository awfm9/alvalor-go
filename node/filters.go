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

package node

import "github.com/alvalor/alvalor-go/types"

//MsgType enum.
type MsgType uint16

const (
	transaction MsgType = 1
)

//MsgFilter filter for Received message type
func MsgFilter(msgType MsgType, hashes ...types.Hash) func(interface{}) bool {
	return func(msg interface{}) bool {
		switch message := msg.(type) {
		case *Transaction:
			return msgType == transaction && len(hashes) == 0 || contains(hashes, message.hash)
		default:
			return false
		}
	}
}

//AnyMsgFilter filter for any message type
func AnyMsgFilter(hashes ...types.Hash) func(interface{}) bool {
	return func(msg interface{}) bool {
		return MsgFilter(transaction, hashes...)(msg) //Add more once you add new message types
	}
}

func contains(hashes []types.Hash, hash types.Hash) bool {
	for _, hsh := range hashes {
		if hsh == hash {
			return true
		}
	}
	return false
}

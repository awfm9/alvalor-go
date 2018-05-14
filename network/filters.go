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

package network

import "time"

func isNot(addresses []string) func(string) bool {
	lookup := make(map[string]struct{})
	for _, address := range addresses {
		lookup[address] = struct{}{}
	}
	return func(address string) bool {
		_, ok := lookup[address]
		return !ok
	}
}

func isScoreAbove(rep reputationManager, threshold float32) func(string) bool {
	return func(address string) bool {
		return rep.Score(address) > threshold
	}
}

func isFailBefore(rep reputationManager, cutoff time.Time) func(string) bool {
	return func(address string) bool {
		return rep.Fail(address).Before(cutoff)
	}
}

//MsgType enum.
type MsgType uint16

const (
	connected    MsgType = 1
	disconnected MsgType = 2
	received     MsgType = 3
)

//MsgFilter can be used to filter out message types.
func MsgFilter(msgType MsgType, addresses ...string) func(interface{}) bool {
	return func(msg interface{}) bool {
		switch message := msg.(type) {
		case *Connected:
			return msgType == connected && len(addresses) == 0 || contains(addresses, message.Address)
		case *Disconnected:
			return msgType == disconnected && len(addresses) == 0 || contains(addresses, message.Address)
		case *Received:
			return msgType == received && len(addresses) == 0 || contains(addresses, message.Address)
		default:
			return false
		}
	}
}

//AnyMsgFilter filter for any message type
func AnyMsgFilter(addresses ...string) func(interface{}) bool {
	return func(msg interface{}) bool {
		return MsgFilter(connected, addresses...)(msg) || MsgFilter(disconnected, addresses...)(msg) || MsgFilter(received, addresses...)(msg)
	}
}

func contains(addresses []string, addr string) bool {
	for _, address := range addresses {
		if address == addr {
			return true
		}
	}
	return false
}

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

//DisconnectedMsgFilter filter for Disconnected message type
func DisconnectedMsgFilter(addresses ...string) func(interface{}) bool {
	return func(msg interface{}) bool {
		switch message := msg.(type) {
		case *Disconnected:
			return len(addresses) == 0 || contains(addresses, message.Address)
		default:
			return false
		}
	}
}

//ConnectedMsgFilter filter for Connected message type
func ConnectedMsgFilter(addresses ...string) func(interface{}) bool {
	return func(msg interface{}) bool {
		switch message := msg.(type) {
		case *Connected:
			return len(addresses) == 0 || contains(addresses, message.Address)
		default:
			return false
		}
	}
}

//ReceivedMsgFilter filter for Received message type
func ReceivedMsgFilter(addresses ...string) func(interface{}) bool {
	return func(msg interface{}) bool {
		switch message := msg.(type) {
		case *Received:
			return len(addresses) == 0 || contains(addresses, message.Address)
		default:
			return false
		}
	}
}

//AnyMsgFilter filter for any message type
func AnyMsgFilter(addresses ...string) func(interface{}) bool {
	return func(msg interface{}) bool {
		return ConnectedMsgFilter(addresses...)(msg) || DisconnectedMsgFilter(addresses...)(msg) || ReceivedMsgFilter(addresses...)(msg)
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

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

// Protocol represents a number of rules on how to handle messages.
type Protocol interface {
	Process(Message, State, Peer) ([]Message, []string, error)
}

// VersionOne represents the first version of the network protocol.
type VersionOne struct {
}

// Process processes a message with this protocol version.
func (v VersionOne) Process(msg Message, state State, peer Peer) ([]Message, []string, error) {
	return nil, nil, nil
}

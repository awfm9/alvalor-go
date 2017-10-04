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

import "net"

// Balance represents a request to add or remove peers.
type Balance struct {
	Min int
	Max int
}

// Failure represents a connection failure event on a given address.
type Failure struct {
	Address string
}

// Violation represents an invalid peer on a given address.
type Violation struct {
	Address string
}

// Error represens a messaging error on a given address.
type Error struct {
	Address string
}

// Message represents a message event on a given address.
type Message struct {
	Address string
	Value   interface{}
}

// Connection represents a connection event on a given address.
type Connection struct {
	Address string
	Conn    net.Conn
	Nonce   []byte
}

// Disconnection represents a disconnection event on a given address.
type Disconnection struct {
	Address string
}

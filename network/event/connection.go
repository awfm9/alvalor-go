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

package event

import "net"

// Enum of possible events on network connections.
const (
	ConnectionIncoming = iota
	ConnectionOutgoing
	ConnectionEstablished
	ConnectionError
	ConnectionFailed
)

// Connection describes an event that can happen on network connections.
type Connection struct {
	Type uint8
	Conn net.Conn
	Err  error
}

// Address returns the string address of the underlying connection.
func (c Connection) Address() string {
	return c.Conn.RemoteAddr().String()
}

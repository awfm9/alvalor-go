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

import (
	"net"
	"time"

	"github.com/pkg/errors"
)

// Listener defines our own listener as the standard library interface doesn't have the deadline functions.
type Listener interface {
	Accept() (net.Conn, error)
	Close() error
	SetDeadline(t time.Time) error
}

type listenManager interface {
	Listen(address string) (Listener, error)
}

type simpleListenManager struct{}

func (lm simpleListenManager) Listen(address string) (Listener, error) {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return nil, errors.Wrap(err, "could not listen")
	}
	return ln.(*net.TCPListener), nil
}

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
	"github.com/stretchr/testify/mock"
	"net"
	"time"
)

type connMock struct {
	mock.Mock
}

func (conn *connMock) Read(b []byte) (n int, err error) {
	args := conn.Called(b)
	return args.Int(0), args.Error(1)
}
func (conn *connMock) Write(b []byte) (n int, err error) {
	args := conn.Called(b)
	return args.Int(0), args.Error(1)
}
func (conn *connMock) Close() error {
	args := conn.Called()
	return args.Error(0)
}
func (conn *connMock) LocalAddr() net.Addr {
	args := conn.Called()
	result, _ := args.Get(0).(*addrMock)
	return result
}
func (conn *connMock) RemoteAddr() net.Addr {
	args := conn.Called()
	result, _ := args.Get(0).(*addrMock)
	return result
}
func (conn *connMock) SetDeadline(t time.Time) error {
	args := conn.Called()
	return args.Error(0)
}
func (conn *connMock) SetReadDeadline(t time.Time) error {
	args := conn.Called()
	return args.Error(0)
}
func (conn *connMock) SetWriteDeadline(t time.Time) error {
	args := conn.Called()
	return args.Error(0)
}

type addrMock struct {
	mock.Mock
}

func (addr *addrMock) Network() string {
	args := addr.Called()
	return args.String(0)
}
func (addr *addrMock) String() string {
	args := addr.Called()
	return args.String(0)
}

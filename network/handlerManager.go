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
	"io"
	"net"
)

type handlerManager interface {
	Accept(conn net.Conn)
	Connect()
	Listen()
	Send(address string, output <-chan interface{}, w io.Writer)
	Process(address string, input <-chan interface{}, output chan<- interface{})
	Receive(address string, r io.Reader, input chan<- interface{})
}

type simpleHandlerManager struct{}

// TODO: implement constructor which injects all dependencies

// TODO: implement the actual handler launching

func (hm simpleHandlerManager) Accept(conn net.Conn) {}

func (hm simpleHandlerManager) Connect() {}

func (hm simpleHandlerManager) Listen() {}

func (hm simpleHandlerManager) Send(address string, output <-chan interface{}, w io.Writer) {}

func (hm simpleHandlerManager) Process(address string, input <-chan interface{}, output chan<- interface{}) {
}

func (hm simpleHandlerManager) Receive(address string, r io.Reader, input chan<- interface{}) {}

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

	"github.com/pierrec/lz4"
	"github.com/pkg/errors"
)

// Sender manages all the output channels to send messages to peers.
type Sender struct {
	outputs map[string]chan<- interface{}
}

func (s *Sender) addOutput(address string, codec Codec, conn net.Conn) error {
	_, ok := s.outputs[address]
	if ok {
		return errors.Errorf("output already exists: %v", address)
	}
	writer := lz4.NewWriter(conn)
	output := make(chan interface{})
	s.outputs[address] = output
	go handleOutgoing(output, codec, writer)
	return nil
}

// Send will try to deliver the message to the peer with the address.
func (s *Sender) Send(address string, message interface{}) error {
	c, ok := s.outputs[address]
	if !ok {
		return errors.Errorf("output doesn't exist: %v", address)
	}
	select {
	case c <- message:
		return nil
	default:
		return errors.Errorf("output timed out: %v", address)
	}
}

func handleOutgoing(output <-chan interface{}, codec Codec, writer io.Writer) {
	for msg := range output {
		err := codec.Encode(writer, msg)
		if err != nil {
			// TODO: handle the error somehow
			break
		}
	}
}

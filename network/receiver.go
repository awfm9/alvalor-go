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

// Receiver is responsible for receiving and multiplexing all message.
type Receiver struct {
	inputs map[string]<-chan interface{}
}

// addInput adds a new input to the receiver to read messages from.
func (r *Receiver) addInput(address string, codec Codec, conn net.Conn) error {
	_, ok := r.inputs[address]
	if ok {
		return errors.Errorf("input already exists: %v", address)
	}
	reader := lz4.NewReader(conn)
	input := make(chan interface{})
	r.inputs[address] = input
	go handleInput(reader, codec, input)
	return nil
}

func handleInput(reader io.Reader, codec Codec, input chan<- interface{}) {
	for {
		msg, err := codec.Decode(reader)
		if err != nil {
			// TODO: handle error
			break
		}
		input <- msg
	}
}

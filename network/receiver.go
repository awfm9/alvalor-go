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
	"github.com/rs/zerolog"
)

// Receiver is responsible for receiving and multiplexing all message.
type Receiver struct {
	log    zerolog.Logger
	input  chan<- interface{}
	inputs map[string]net.Conn
}

// NewReceiver creates a new receiver with the given input channel as feed for
// new received messages.
func NewReceiver(log zerolog.Logger, input chan<- interface{}) *Receiver {
	return &Receiver{
		log:    log,
		input:  input,
		inputs: make(map[string]net.Conn),
	}
}

func (r *Receiver) addInput(address string, codec Codec, conn net.Conn) error {
	_, ok := r.inputs[address]
	if ok {
		return errors.Errorf("input already exists: %v", address)
	}
	r.inputs[address] = conn
	go handleReceiving(r.log, codec, conn, r.input)
	return nil
}

func (r *Receiver) removeInput(address string) error {
	conn, ok := r.inputs[address]
	if !ok {
		return errors.Errorf("input not found: %v", address)
	}
	defer delete(r.inputs, address)
	err := conn.Close()
	if err != nil {
		return errors.Wrap(err, "could not close connection")
	}
	return nil
}

func handleReceiving(log zerolog.Logger, codec Codec, conn net.Conn, input chan<- interface{}) {
	address := conn.RemoteAddr().String()
	reader := lz4.NewReader(conn)
	for {
		msg, err := codec.Decode(reader)
		if err != nil && err == io.EOF {
			log.Info().Str("address", address).Msg("network connection closed")
			break
		}
		if err != nil {
			log.Error().Str("address", address).Err(err).Msg("reading message failed")
			continue
		}
		input <- msg
	}
}

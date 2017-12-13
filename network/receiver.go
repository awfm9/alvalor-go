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
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Receiver is responsible for receiving and multiplexing all message.
type Receiver struct {
	log    zerolog.Logger
	input  chan<- interface{}
	inputs map[string]net.Conn
	wg     *sync.WaitGroup
}

// NewReceiver creates a new receiver with the given input channel as feed for
// new received messages.
func NewReceiver(log zerolog.Logger, input chan<- interface{}) *Receiver {
	return &Receiver{
		log:    log,
		input:  input,
		inputs: make(map[string]net.Conn),
		wg:     &sync.WaitGroup{},
	}
}

func (r *Receiver) addInput(address string, codec Codec, conn net.Conn) error {
	_, ok := r.inputs[address]
	if ok {
		return errors.Errorf("input already exists (%v)", address)
	}
	r.inputs[address] = conn
	r.wg.Add(1)
	go handleReceiving(r.log, r.wg, codec, conn, r.input)
	return nil
}

func (r *Receiver) removeInput(address string) error {
	conn, ok := r.inputs[address]
	if !ok {
		return errors.Errorf("input not found (%v)", address)
	}
	conn.Close()
	delete(r.inputs, address)
	return nil
}

func (r *Receiver) stop() {
	for _, conn := range r.inputs {
		conn.Close()
	}
	r.wg.Wait()
}

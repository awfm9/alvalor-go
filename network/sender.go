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

// Sender manages all the output channels to send messages to peers.
type Sender struct {
	log        zerolog.Logger
	outputs    map[string]chan<- interface{}
	bufferSize uint
	wg         *sync.WaitGroup
}

// NewSender creates a new sender responsible for sending messages to peers.
func NewSender(log zerolog.Logger) *Sender {
	return &Sender{
		log:        log,
		outputs:    make(map[string]chan<- interface{}),
		bufferSize: 16,
		wg:         &sync.WaitGroup{},
	}
}

func (s *Sender) addOutput(address string, codec Codec, conn net.Conn) error {
	_, ok := s.outputs[address]
	if ok {
		return errors.Errorf("output already exists: %v", address)
	}
	output := make(chan interface{}, s.bufferSize)
	s.outputs[address] = output
	s.wg.Add(1)
	go handleSending(s.log, s.wg, output, codec, conn)
	return nil
}

func (s *Sender) removeOutput(address string) error {
	output, ok := s.outputs[address]
	if !ok {
		return errors.Errorf("output not found (%v)", address)
	}
	close(output)
	delete(s.outputs, address)
	return nil
}

// Send will try to deliver the message to the peer with the address.
func (s *Sender) Send(address string, message interface{}) error {
	output, ok := s.outputs[address]
	if !ok {
		return errors.Errorf("output doesn't exist (%v)", address)
	}
	select {
	case output <- message:
		return nil
	default:
		return errors.Errorf("output timed out (%v)", address)
	}
}

func (s *Sender) stop() {
	for _, output := range s.outputs {
		close(output)
	}
	s.wg.Wait()
}

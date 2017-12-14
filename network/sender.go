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

	"github.com/pierrec/lz4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Sender manages all the output channels to send messages to peers.
type Sender struct {
	log        zerolog.Logger
	wg         *sync.WaitGroup
	outputs    map[string]chan interface{}
	bufferSize uint
}

// NewSender creates a new sender responsible for sending messages to peers.
func NewSender(log zerolog.Logger) *Sender {
	return &Sender{
		log:        log,
		outputs:    make(map[string]chan interface{}),
		bufferSize: 16,
		wg:         &sync.WaitGroup{},
	}
}

func (s *Sender) addOutput(conn net.Conn, codec Codec) (chan<- interface{}, error) {
	address := conn.RemoteAddr().String()
	_, ok := s.outputs[address]
	if ok {
		return nil, errors.Errorf("output already exists: %v", address)
	}
	output := make(chan interface{}, s.bufferSize)
	s.outputs[address] = output
	s.wg.Add(1)
	go handleSending(s.log, s.wg, output, codec, conn)
	return output, nil
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

func (s *Sender) stop() {
	for address, output := range s.outputs {
		close(output)
		delete(s.outputs, address)
	}
	s.wg.Wait()
}

func handleSending(log zerolog.Logger, wg *sync.WaitGroup, output <-chan interface{}, codec Codec, conn net.Conn) {
	defer wg.Done()
	address := conn.RemoteAddr().String()
	log = log.With().Str("component", "sender").Str("address", address).Logger()
	log.Info().Msg("sending routine started")
	defer log.Info().Msg("sending routine stopped")
	writer := lz4.NewWriter(conn)
	for msg := range output {
		err := codec.Encode(writer, msg)
		if err != nil {
			log.Error().Err(err).Msg("could not write message")
			continue
		}
	}
}

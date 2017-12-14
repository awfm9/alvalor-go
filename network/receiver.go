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
	"sync"

	"github.com/pierrec/lz4"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Receiver is responsible for receiving and multiplexing all message.
type Receiver struct {
	log        zerolog.Logger
	wg         *sync.WaitGroup
	inputs     map[string]chan interface{}
	bufferSize uint
}

// NewReceiver creates a new receiver with the given input channel as feed for
// new received messages.
func NewReceiver(log zerolog.Logger) *Receiver {
	return &Receiver{
		log:        log,
		inputs:     make(map[string]chan interface{}),
		wg:         &sync.WaitGroup{},
		bufferSize: 16,
	}
}

func (r *Receiver) addInput(conn net.Conn, codec Codec) (<-chan interface{}, error) {
	address := conn.RemoteAddr().String()
	_, ok := r.inputs[address]
	if ok {
		return nil, errors.Errorf("input already exists (%v)", address)
	}
	input := make(chan interface{}, r.bufferSize)
	r.inputs[address] = input
	r.wg.Add(1)
	go handleReceiving(r.log, r.wg, codec, conn, input)
	return input, nil
}

func (r *Receiver) removeInput(address string) error {
	input, ok := r.inputs[address]
	if !ok {
		return errors.Errorf("input not found (%v)", address)
	}
	close(input)
	delete(r.inputs, address)
	return nil
}

func (r *Receiver) stop() {
	for address, input := range r.inputs {
		close(input)
		delete(r.inputs, address)
	}
	r.wg.Wait()
}

func handleReceiving(log zerolog.Logger, wg *sync.WaitGroup, codec Codec, conn net.Conn, input chan<- interface{}) {
	defer wg.Done()

	// extract configuration as needed
	var (
		address = conn.RemoteAddr().String()
	)

	// configure logger and add start/stop messages
	log = log.With().Str("component", "receiver").Str("address", address).Logger()
	log.Info().Msg("receiving routine started")
	defer log.Info().Msg("receiving routine closed")

	// read all messages from connetion and forward on channel
	reader := lz4.NewReader(conn)
	for {
		msg, err := codec.Decode(reader)
		if err != nil && err == io.EOF {
			log.Info().Msg("network connection closed")
			break
		}
		if err != nil {
			log.Error().Err(err).Msg("reading message failed")
			continue
		}
		input <- msg
	}
}

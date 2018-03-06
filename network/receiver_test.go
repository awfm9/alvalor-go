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
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/awishformore/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestReceiver(t *testing.T) {
	suite.Run(t, new(ReceiverSuite))
}

type ReceiverSuite struct {
	suite.Suite
	log zerolog.Logger
	cfg Config
	wg  sync.WaitGroup
}

func (suite *ReceiverSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.cfg = Config{}
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
}

func (suite *ReceiverSuite) TestReceiverSuccess() {

	// arrange
	address := "192.0.2.100:1337"
	input := make(chan interface{}, 16)
	r := &bytes.Buffer{}

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)

	codec := &CodecMock{}
	codec.On("Decode", r).Return(&Ping{}, nil).Once()
	codec.On("Decode", r).Return(&Pong{}, nil).Once()
	codec.On("Decode", r).Return(&Discover{}, nil).Once()
	codec.On("Decode", r).Return(&Peers{}, nil).Once()
	codec.On("Decode", r).Return(nil, io.EOF)

	peers := &PeerManagerMock{}
	peers.On("Drop", mock.Anything).Return(nil)

	// act
	suite.cfg.codec = codec
	go handleReceiving(suite.log, &suite.wg, &suite.cfg, rep, peers, address, r, input)
	var msgs []interface{}
	for msg := range input {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	t := suite.T()

	if assert.Len(t, msgs, 4) {
		assert.IsType(t, &Ping{}, msgs[0])
		assert.IsType(t, &Pong{}, msgs[1])
		assert.IsType(t, &Discover{}, msgs[2])
		assert.IsType(t, &Peers{}, msgs[3])
	}

	peers.AssertCalled(t, "Drop", address)

	rep.AssertNotCalled(t, "Failure", mock.Anything)
}

func (suite *ReceiverSuite) TestReceiverEOF() {

	// arrange
	address := "192.0.2.100:1337"
	input := make(chan interface{}, 16)
	r := &bytes.Buffer{}

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)

	codec := &CodecMock{}
	codec.On("Decode", r).Return(nil, io.EOF)

	peers := &PeerManagerMock{}
	peers.On("Drop", mock.Anything).Return(nil)

	// act
	suite.cfg.codec = codec
	go handleReceiving(suite.log, &suite.wg, &suite.cfg, rep, peers, address, r, input)
	suite.wg.Wait()

	// assert
	t := suite.T()

	_, ok := <-input
	assert.False(t, ok)

	peers.AssertCalled(t, "Drop", address)

	rep.AssertNotCalled(t, "Failure", mock.Anything)
}

func (suite *ReceiverSuite) TestReceiverError() {

	// arrange
	address := "192.0.2.100:1337"
	input := make(chan interface{}, 16)
	r := &bytes.Buffer{}

	message := "message"

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)

	codec := &CodecMock{}
	codec.On("Decode", r).Return(nil, errors.New("could not encode message")).Once()
	codec.On("Decode", r).Return(message, nil).Once()
	codec.On("Decode", r).Return(nil, io.EOF)

	peers := &PeerManagerMock{}
	peers.On("Drop", mock.Anything).Return(nil)

	// act
	suite.cfg.codec = codec
	go handleReceiving(suite.log, &suite.wg, &suite.cfg, rep, peers, address, r, input)
	var msgs []interface{}
	for msg := range input {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	t := suite.T()

	if assert.Len(t, msgs, 1) {
		assert.Equal(t, message, msgs[0])
	}

	rep.AssertCalled(t, "Failure", address)
	peers.AssertCalled(t, "Drop", address)
}

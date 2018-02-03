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

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestReceiver(t *testing.T) {
	suite.Run(t, new(ReceiverSuite))
}

type ReceiverSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *ReceiverSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{}
}

func (suite *ReceiverSuite) TestReceiverEOFError() {

	// arrange
	address := "15.77.14.74:5454"
	input := make(chan interface{}, 16)
	r := &bytes.Buffer{}

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}

	codec := &CodecMock{}
	codec.On("Decode", r).Return(nil, io.EOF)

	// act
	suite.cfg.codec = codec
	go handleReceiving(suite.log, &suite.wg, &suite.cfg, peers, rep, address, r, input)
	suite.wg.Wait()

	// assert
	_, ok := <-input
	assert.False(suite.T(), ok)
}

func (suite *ReceiverSuite) TestReceiverClosedError() {

	// arrange
	address := "15.77.14.74:5454"
	input := make(chan interface{}, 16)
	r := &bytes.Buffer{}

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}

	codec := &CodecMock{}
	codec.On("Decode", r).Return(nil, errors.New("use of closed network connection"))

	// act
	suite.cfg.codec = codec
	go handleReceiving(suite.log, &suite.wg, &suite.cfg, peers, rep, address, r, input)
	suite.wg.Wait()

	// assert
	_, ok := <-input
	assert.False(suite.T(), ok)
}

func (suite *ReceiverSuite) TestReceiverReceiveMessages() {

	// arrange
	address := "15.77.14.74:5454"
	input := make(chan interface{}, 16)
	r := &bytes.Buffer{}

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}

	codec := &CodecMock{}
	codec.On("Decode", r).Return(&Ping{}, nil).Once()
	codec.On("Decode", r).Return(&Pong{}, nil).Once()
	codec.On("Decode", r).Return(&Discover{}, nil).Once()
	codec.On("Decode", r).Return(&Peers{}, nil).Once()
	codec.On("Decode", r).Return(nil, io.EOF)

	// act
	suite.cfg.codec = codec
	go handleReceiving(suite.log, &suite.wg, &suite.cfg, peers, rep, address, r, input)
	var msgs []interface{}
	for msg := range input {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	if assert.Len(suite.T(), msgs, 4) {
		assert.IsType(suite.T(), &Ping{}, msgs[0])
		assert.IsType(suite.T(), &Pong{}, msgs[1])
		assert.IsType(suite.T(), &Discover{}, msgs[2])
		assert.IsType(suite.T(), &Peers{}, msgs[3])
	}
}

func (suite *ReceiverSuite) TestReceiverDecodeFails() {

	// arrange
	address := "15.77.14.74:5454"
	message := "some message"
	input := make(chan interface{}, 16)
	r := &bytes.Buffer{}

	peers := &PeerManagerMock{}
	peers.On("Drop", address).Return(errors.New("dropping failed"))

	rep := &ReputationManagerMock{}
	rep.On("Error", address)

	codec := &CodecMock{}
	codec.On("Decode", r).Return(nil, errors.New("could not encode message")).Once()
	codec.On("Decode", r).Return(message, nil).Once()
	codec.On("Decode", r).Return(nil, io.EOF)

	// act
	suite.cfg.codec = codec
	go handleReceiving(suite.log, &suite.wg, &suite.cfg, peers, rep, address, r, input)
	var msgs []interface{}
	for msg := range input {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	rep.AssertCalled(suite.T(), "Error", address)
	peers.AssertCalled(suite.T(), "Drop", address)
	if assert.Len(suite.T(), msgs, 1) {
		assert.Equal(suite.T(), message, msgs[0])
	}
}

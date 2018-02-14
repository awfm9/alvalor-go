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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestSender(t *testing.T) {
	suite.Run(t, new(SenderSuite))
}

type SenderSuite struct {
	suite.Suite
	log zerolog.Logger
	cfg Config
	wg  sync.WaitGroup
}

func (suite *SenderSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.cfg = Config{}
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
}

func (suite *SenderSuite) TestSenderEOFError() {

	// arrange
	address := "192.0.2.100:1337"
	output := make(chan interface{}, 16)
	w := &bytes.Buffer{}

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}

	codec := &CodecMock{}
	codec.On("Encode", w, mock.Anything).Return(io.EOF).Once()
	codec.On("Encode", w, mock.Anything).Return(nil)

	eventMgr := &EventManagerMock{}
	eventMgr.On("Disconnected", mock.Anything)

	// act
	suite.cfg.codec = codec
	go handleSending(suite.log, &suite.wg, &suite.cfg, peers, rep, address, output, w, eventMgr)
	output <- &Ping{}
	output <- &Pong{}
	output <- &Discover{}
	close(output)
	suite.wg.Wait()

	// assert
	if codec.AssertNumberOfCalls(suite.T(), "Encode", 1) {
		codec.AssertCalled(suite.T(), "Encode", w, &Ping{})
	}
}

func (suite *SenderSuite) TestSenderSendMessages() {

	// arrange
	address := "192.0.2.100:1337"
	output := make(chan interface{}, 16)
	w := &bytes.Buffer{}

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}

	codec := &CodecMock{}
	codec.On("Encode", w, mock.Anything).Return(nil)

	eventMgr := &EventManagerMock{}
	eventMgr.On("Disconnected", mock.Anything)

	// act
	suite.cfg.codec = codec
	go handleSending(suite.log, &suite.wg, &suite.cfg, peers, rep, address, output, w, eventMgr)
	output <- &Ping{}
	output <- &Pong{}
	output <- &Discover{}
	output <- &Peers{}
	close(output)
	suite.wg.Wait()

	// assert
	if codec.AssertNumberOfCalls(suite.T(), "Encode", 4) {
		codec.AssertCalled(suite.T(), "Encode", w, &Ping{})
		codec.AssertCalled(suite.T(), "Encode", w, &Pong{})
		codec.AssertCalled(suite.T(), "Encode", w, &Discover{})
		codec.AssertCalled(suite.T(), "Encode", w, &Peers{})
	}
}

func (suite *SenderSuite) TestSenderEncodeFails() {

	// arrange
	address := "192.0.2.100:1337"
	output := make(chan interface{}, 16)
	w := &bytes.Buffer{}

	peers := &PeerManagerMock{}
	peers.On("Drop", address).Return(errors.New("could not drop peer"))

	rep := &ReputationManagerMock{}
	rep.On("Error", address)

	codec := &CodecMock{}
	codec.On("Encode", w, mock.Anything).Return(errors.New("could not decode message"))
	codec.On("Encode", w, mock.Anything).Return(nil).Twice()
	codec.On("Encode", w, mock.Anything).Return(io.EOF)

	eventMgr := &EventManagerMock{}
	eventMgr.On("Disconnected", mock.Anything)

	// act
	suite.cfg.codec = codec
	go handleSending(suite.log, &suite.wg, &suite.cfg, peers, rep, address, output, w, eventMgr)
	output <- &Ping{}
	output <- &Pong{}
	output <- &Discover{}
	output <- &Peers{}
	close(output)
	suite.wg.Wait()

	// assert
	rep.AssertCalled(suite.T(), "Error", address)
	peers.AssertCalled(suite.T(), "Drop", address)
	if codec.AssertNumberOfCalls(suite.T(), "Encode", 4) {
		codec.AssertCalled(suite.T(), "Encode", w, &Ping{})
		codec.AssertCalled(suite.T(), "Encode", w, &Pong{})
		codec.AssertCalled(suite.T(), "Encode", w, &Discover{})
		codec.AssertCalled(suite.T(), "Encode", w, &Peers{})
	}
}

func (suite *SenderSuite) TestSenderEncodeFailsAndDisconnectedPublished() {

	// arrange
	address := "192.0.2.100:1337"
	output := make(chan interface{}, 16)
	w := &bytes.Buffer{}

	peers := &PeerManagerMock{}
	peers.On("Drop", address).Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Error", address)

	codec := &CodecMock{}
	codec.On("Encode", w, mock.Anything).Return(errors.New("could not decode message"))

	eventMgr := &EventManagerMock{}
	eventMgr.On("Disconnected", mock.Anything)

	// act
	suite.cfg.codec = codec
	go handleSending(suite.log, &suite.wg, &suite.cfg, peers, rep, address, output, w, eventMgr)
	output <- &Ping{}
	output <- &Pong{}
	output <- &Discover{}
	output <- &Peers{}
	close(output)
	suite.wg.Wait()

	// assert
	eventMgr.AssertCalled(suite.T(), "Disconnected", address)
}

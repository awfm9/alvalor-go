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
	"time"

	"github.com/awishformore/zerolog"
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
	suite.cfg = Config{
		interval: 2 * time.Millisecond,
	}
}

func (suite *SenderSuite) TestSenderSuccess() {

	// arrange
	address := "192.0.2.100:1337"
	output := make(chan interface{}, 5)
	w := &bytes.Buffer{}

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)

	codec := &CodecMock{}
	codec.On("Encode", mock.Anything, mock.Anything).Return(nil)

	events := &EventManagerMock{}
	events.On("Disconnected", mock.Anything).Return(nil)

	// act
	suite.cfg.codec = codec
	go handleSending(suite.log, &suite.wg, &suite.cfg, rep, events, address, output, w)
	output <- &Ping{}
	output <- &Pong{}
	output <- &Discover{}
	output <- &Peers{}
	close(output)
	suite.wg.Wait()

	// assert
	t := suite.T()

	if codec.AssertNumberOfCalls(t, "Encode", 4) {
		codec.AssertCalled(t, "Encode", w, &Ping{})
		codec.AssertCalled(t, "Encode", w, &Pong{})
		codec.AssertCalled(t, "Encode", w, &Discover{})
		codec.AssertCalled(t, "Encode", w, &Peers{})
	}
}

func (suite *SenderSuite) TestSenderEOF() {

	// arrange
	address := "192.0.2.100:1337"
	output := make(chan interface{}, 5)
	w := &bytes.Buffer{}

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)

	codec := &CodecMock{}
	codec.On("Encode", mock.Anything, mock.Anything).Return(io.EOF)
	codec.On("Encode", mock.Anything, mock.Anything).Return(nil)

	events := &EventManagerMock{}
	events.On("Disconnected", mock.Anything).Return(nil)

	// act
	suite.cfg.codec = codec
	go handleSending(suite.log, &suite.wg, &suite.cfg, rep, events, address, output, w)
	output <- &Ping{}
	output <- &Pong{}
	output <- &Discover{}
	output <- &Peers{}
	close(output)
	suite.wg.Wait()

	// assert
	if codec.AssertNumberOfCalls(suite.T(), "Encode", 1) {
		codec.AssertCalled(suite.T(), "Encode", w, &Ping{})
	}
}

func (suite *SenderSuite) TestSenderHeartbeat() {

	// arrange
	address := "192.0.2.100:1337"
	output := make(chan interface{}, 5)
	w := &bytes.Buffer{}

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)

	codec := &CodecMock{}
	codec.On("Encode", mock.Anything, mock.Anything).Return(nil)

	events := &EventManagerMock{}
	events.On("Disconnected", mock.Anything).Return(nil)

	// act
	suite.cfg.codec = codec
	go handleSending(suite.log, &suite.wg, &suite.cfg, rep, events, address, output, w)
	time.Sleep(time.Duration(1.5 * float64(suite.cfg.interval)))
	close(output)
	suite.wg.Wait()

	// assert
	t := suite.T()

	if codec.AssertNumberOfCalls(t, "Encode", 1) {
		codec.AssertCalled(t, "Encode", w, &Ping{})
	}
}

func (suite *SenderSuite) TestSenderEncodeFails() {

	// arrange
	address := "192.0.2.100:1337"
	output := make(chan interface{}, 5)
	w := &bytes.Buffer{}

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)

	codec := &CodecMock{}
	codec.On("Encode", mock.Anything, mock.Anything).Return(errors.New("could not encode message")).Once()
	codec.On("Encode", mock.Anything, mock.Anything).Return(nil).Once()
	codec.On("Encode", mock.Anything, mock.Anything).Return(io.EOF)

	events := &EventManagerMock{}
	events.On("Disconnected", mock.Anything).Return(nil)

	// act
	suite.cfg.codec = codec
	go handleSending(suite.log, &suite.wg, &suite.cfg, rep, events, address, output, w)
	output <- &Ping{}
	output <- &Pong{}
	output <- &Discover{}
	output <- &Peers{}
	close(output)
	suite.wg.Wait()

	// assert
	rep.AssertCalled(suite.T(), "Failure", address)
	events.AssertCalled(suite.T(), "Disconnected", address)
	if codec.AssertNumberOfCalls(suite.T(), "Encode", 3) {
		codec.AssertCalled(suite.T(), "Encode", w, &Ping{})
		codec.AssertCalled(suite.T(), "Encode", w, &Pong{})
		codec.AssertCalled(suite.T(), "Encode", w, &Discover{})
	}
}

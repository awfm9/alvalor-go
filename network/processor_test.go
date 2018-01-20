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
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ProcessorTestSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *ProcessorTestSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		interval: 25 * time.Millisecond,
		address:  "153.66.22.77:5412",
	}
}

func (suite *ProcessorTestSuite) TestProcessingPublishesPeers() {
	// arrange
	infos := &processorInfosMock{}
	actions := &processorActionsMock{}
	events := &processorEventsMock{}
	address := "15.77.14.74:5454"
	input := make(chan interface{}, 1)
	output := make(chan interface{})
	stop := make(chan struct{})
	suite.cfg.listen = true

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, infos, actions, events, address, input, output)
	msg := <-output

	// assert
	assert.IsType(suite.T(), &Peers{}, msg)
	peersMsg := msg.(*Peers)
	assert.EqualValues(suite.T(), []string{suite.cfg.address}, peersMsg.Addresses)
}

func (suite *ProcessorTestSuite) TestProcessingPublishesDiscover() {
	// arrange
	infos := &processorInfosMock{}
	actions := &processorActionsMock{}
	events := &processorEventsMock{}
	address := "15.77.14.74:5454"
	input := make(chan interface{}, 1)
	output := make(chan interface{})
	stop := make(chan struct{})
	suite.cfg.listen = false

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, infos, actions, events, address, input, output)
	msg := <-output

	// assert
	assert.IsType(suite.T(), &Discover{}, msg)
}

func (suite *ProcessorTestSuite) TestProcessingPublishesPing() {
	// arrange
	infos := &processorInfosMock{}
	actions := &processorActionsMock{}
	actions.On("DropPeer", mock.Anything).Return(nil)
	events := &processorEventsMock{}

	address := "15.77.14.74:5454"
	input := make(chan interface{}, 1)
	output := make(chan interface{}, 5)
	stop := make(chan struct{})
	suite.cfg.listen = false

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, infos, actions, events, address, input, output)
	time.Sleep(50 * time.Millisecond)

	// assert
	msg := <-output
	msg = <-output
	assert.IsType(suite.T(), &Ping{}, msg)
}

func (suite *ProcessorTestSuite) TestProcessingPublishesPong() {
	// arrange
	infos := &processorInfosMock{}
	actions := &processorActionsMock{}
	events := &processorEventsMock{}
	address := "15.77.14.74:5454"
	input := make(chan interface{}, 1)
	output := make(chan interface{})
	stop := make(chan struct{})
	suite.cfg.listen = false

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	input <- &Ping{}
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, infos, actions, events, address, input, output)
	msg := <-output
	msg = <-output

	// assert
	assert.IsType(suite.T(), &Pong{}, msg)
}

func (suite *ProcessorTestSuite) TestProcessingPublishesPeersIfDiscoverReceived() {
	// arrange
	infos := &processorInfosMock{}
	addresses := []string{"17.63.23.55:5345", "88.22.77.55:3442"}
	infos.On("AddressSample").Return(addresses, nil)

	actions := &processorActionsMock{}
	events := &processorEventsMock{}
	address := "15.77.14.74:5454"
	input := make(chan interface{}, 1)
	output := make(chan interface{})
	stop := make(chan struct{})
	suite.cfg.listen = false

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	input <- &Discover{}
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, infos, actions, events, address, input, output)
	msg := <-output
	msg = <-output

	// assert
	assert.IsType(suite.T(), &Peers{}, msg)
	peersMsg := msg.(*Peers)
	assert.EqualValues(suite.T(), addresses, peersMsg.Addresses)
}

func (suite *ProcessorTestSuite) TestProcessingPublishesFound() {
	// arrange
	infos := &processorInfosMock{}
	actions := &processorActionsMock{}
	actions.On("DropPeer", mock.Anything).Return(nil)
	events := &processorEventsMock{}
	addresses := []string{"17.63.23.55:5345", "88.22.77.55:3442"}
	events.On("Found", addresses[0])
	events.On("Found", addresses[1])

	address := "15.77.14.74:5454"
	input := make(chan interface{}, 1)
	output := make(chan interface{}, 5)
	stop := make(chan struct{})
	suite.cfg.listen = false

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	input <- &Peers{Addresses: addresses}
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, infos, actions, events, address, input, output)
	time.Sleep(50 * time.Millisecond)

	// assert
	events.AssertCalled(suite.T(), "Found", addresses[0])
	events.AssertCalled(suite.T(), "Found", addresses[1])
}

func (suite *ProcessorTestSuite) TestProcessingDropsPeer() {
	// arrange
	infos := &processorInfosMock{}
	actions := &processorActionsMock{}
	actions.On("DropPeer", mock.Anything).Return(nil)
	events := &processorEventsMock{}

	address := "15.77.14.74:5454"
	input := make(chan interface{}, 1)
	output := make(chan interface{}, 5)
	stop := make(chan struct{})
	suite.cfg.listen = false

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, infos, actions, events, address, input, output)
	time.Sleep(100 * time.Millisecond)

	// assert
	actions.AssertCalled(suite.T(), "DropPeer", address)
}

func TestProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(ProcessorTestSuite))
}

type processorInfosMock struct {
	mock.Mock
}

func (infos *processorInfosMock) AddressSample() ([]string, error) {
	args := infos.Called()
	return args.Get(0).([]string), args.Error(1)
}

type processorActionsMock struct {
	mock.Mock
}

func (actions *processorActionsMock) DropPeer(address string) error {
	args := actions.Called(address)
	return args.Error(0)
}

type processorEventsMock struct {
	mock.Mock
}

func (actions *processorEventsMock) Found(address string) {
	actions.Called(address)
}

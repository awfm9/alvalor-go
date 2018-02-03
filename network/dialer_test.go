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
// GNU Affero General Public License for more detailb.
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
	"github.com/stretchr/testify/suite"
)

func TestDialer(t *testing.T) {
	suite.Run(t, new(DialerSuite))
}

type DialerSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *DialerSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		interval: 5 * time.Millisecond,
		minPeers: 5,
		maxPeers: 15,
	}
}

func (suite *DialerSuite) TestDialerSuccess() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(3)

	slots := &SlotManagerMock{}
	slots.On("Pending").Return(5)

	handlers := &HandlerManagerMock{}
	handlers.On("Connect")

	// act
	go handleDialing(suite.log, &suite.wg, &suite.cfg, peers, slots, handlers, stop)
	time.Sleep(50 * time.Millisecond)
	close(stop)
	suite.wg.Wait()

	// assert
	handlers.AssertCalled(suite.T(), "Connect")
}

func (suite *DialerSuite) TestDialerEnoughPeers() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(5)

	slots := &SlotManagerMock{}
	slots.On("Pending").Return(3)

	handlers := &HandlerManagerMock{}
	handlers.On("Connect")

	// act
	go handleDialing(suite.log, &suite.wg, &suite.cfg, peers, slots, handlers, stop)
	time.Sleep(50 * time.Millisecond)
	close(stop)
	suite.wg.Wait()

	// assert
	handlers.AssertNotCalled(suite.T(), "Connect")
}

func (suite *DialerSuite) TestDialerMaximumPendingPeers() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(3)

	slots := &SlotManagerMock{}
	slots.On("Pending").Return(12)

	handlers := &HandlerManagerMock{}
	handlers.On("Connect")

	// act
	go handleDialing(suite.log, &suite.wg, &suite.cfg, peers, slots, handlers, stop)
	time.Sleep(50 * time.Millisecond)
	close(stop)
	suite.wg.Wait()

	// assert
	handlers.AssertNotCalled(suite.T(), "Connect")
}

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
	"github.com/stretchr/testify/mock"
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
		interval: 10 * time.Millisecond,
		minPeers: 5,
		maxPeers: 15,
	}
}

func (suite *DialerSuite) TestDialerSuccess() {

	// arrange
	address := "192.0.2.101:1337"
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(3)
	peers.On("Addresses").Return([]string{})

	pending := &PendingManagerMock{}
	pending.On("Count").Return(5)
	pending.On("Addresses").Return([]string{})

	rep := &ReputationManagerMock{}

	addresses := &AddressManagerMock{}
	addresses.On("Sample", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{address})

	handlers := &HandlerManagerMock{}
	handlers.On("Connect", address)

	// act
	go handleDialing(suite.log, &suite.wg, &suite.cfg, peers, pending, addresses, rep, handlers, stop)
	time.Sleep(15 * time.Millisecond)
	close(stop)
	suite.wg.Wait()

	// assert
	addresses.AssertCalled(suite.T(), "Sample", 1,
		mock.AnythingOfType("func(string) bool"),
		mock.AnythingOfType("func(string) bool"),
		mock.AnythingOfType("func(string, string) bool"),
		mock.AnythingOfType("func(string, string) bool"),
	)
	handlers.AssertCalled(suite.T(), "Connect", address)
}

func (suite *DialerSuite) TestDialerNoAddresses() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(3)
	peers.On("Addresses").Return([]string{})

	pending := &PendingManagerMock{}
	pending.On("Count").Return(5)
	pending.On("Addresses").Return([]string{})

	rep := &ReputationManagerMock{}

	addresses := &AddressManagerMock{}
	addresses.On("Sample", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]string{})

	handlers := &HandlerManagerMock{}
	handlers.On("Connect", mock.Anything)

	// act
	go handleDialing(suite.log, &suite.wg, &suite.cfg, peers, pending, addresses, rep, handlers, stop)
	time.Sleep(15 * time.Millisecond)
	close(stop)
	suite.wg.Wait()

	// assert
	addresses.AssertCalled(suite.T(), "Sample", 1,
		mock.AnythingOfType("func(string) bool"),
		mock.AnythingOfType("func(string) bool"),
		mock.AnythingOfType("func(string, string) bool"),
		mock.AnythingOfType("func(string, string) bool"),
	)
	handlers.AssertNotCalled(suite.T(), "Connect")
}

func (suite *DialerSuite) TestDialerEnoughPeers() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(5)

	pending := &PendingManagerMock{}
	pending.On("Count").Return(3)

	addresses := &AddressManagerMock{}

	rep := &ReputationManagerMock{}

	handlers := &HandlerManagerMock{}
	handlers.On("Connect", mock.Anything)

	// act
	go handleDialing(suite.log, &suite.wg, &suite.cfg, peers, pending, addresses, rep, handlers, stop)
	time.Sleep(15 * time.Millisecond)
	close(stop)
	suite.wg.Wait()

	// assert
	addresses.AssertNotCalled(suite.T(), "Sample")
	handlers.AssertNotCalled(suite.T(), "Connect")
}

func (suite *DialerSuite) TestDialerMaximumPendingPeers() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(3)

	pending := &PendingManagerMock{}
	pending.On("Count").Return(12)

	addresses := &AddressManagerMock{}

	rep := &ReputationManagerMock{}

	handlers := &HandlerManagerMock{}
	handlers.On("Connect", mock.Anything)

	// act
	go handleDialing(suite.log, &suite.wg, &suite.cfg, peers, pending, addresses, rep, handlers, stop)
	time.Sleep(15 * time.Millisecond)
	close(stop)
	suite.wg.Wait()

	// assert
	addresses.AssertNotCalled(suite.T(), "Sample")
	handlers.AssertNotCalled(suite.T(), "Connect")
}

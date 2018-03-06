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

	"github.com/awishformore/zerolog"
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
		interval: 2 * time.Millisecond,
		minPeers: 5,
	}
}

func (suite *DialerSuite) TestDialerSuccess() {

	// arrange
	address := "192.0.2.101:1337"
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(4)
	peers.On("Addresses").Return([]string{})

	pending := &PendingManagerMock{}
	pending.On("Addresses").Return([]string{})

	rep := &ReputationManagerMock{}

	addresses := &AddressManagerMock{}
	addresses.On("Sample", mock.Anything, mock.Anything).Return([]string{address})

	handlers := &HandlerManagerMock{}
	handlers.On("Connector", mock.Anything)
	handlers.On("Discoverer")

	// act
	go handleDialing(suite.log, &suite.wg, &suite.cfg, peers, pending, addresses, rep, handlers, stop)
	time.Sleep(time.Duration(1.5 * float64(suite.cfg.interval)))
	close(stop)
	suite.wg.Wait()

	// assert
	t := suite.T()

	handlers.AssertCalled(t, "Connector", address)

	handlers.AssertNotCalled(t, "Discoverer")
}

func (suite *DialerSuite) TestDialerNoAddresses() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(4)
	peers.On("Addresses").Return([]string{})

	pending := &PendingManagerMock{}
	pending.On("Addresses").Return([]string{})

	rep := &ReputationManagerMock{}

	addresses := &AddressManagerMock{}
	addresses.On("Sample", mock.Anything, mock.Anything).Return([]string{})

	handlers := &HandlerManagerMock{}
	handlers.On("Connector", mock.Anything)
	handlers.On("Discoverer")

	// act
	go handleDialing(suite.log, &suite.wg, &suite.cfg, peers, pending, addresses, rep, handlers, stop)
	time.Sleep(time.Duration(1.5 * float64(suite.cfg.interval)))
	close(stop)
	suite.wg.Wait()

	// assert
	t := suite.T()

	handlers.AssertCalled(t, "Discoverer")

	handlers.AssertNotCalled(t, "Connector", mock.Anything)
}

func (suite *DialerSuite) TestDialerEnoughPeers() {

	// arrange
	address := "192.0.2.101:1337"
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(5)
	peers.On("Addresses").Return([]string{})

	pending := &PendingManagerMock{}
	pending.On("Addresses").Return([]string{})

	rep := &ReputationManagerMock{}

	addresses := &AddressManagerMock{}
	addresses.On("Sample", mock.Anything, mock.Anything).Return([]string{address})

	handlers := &HandlerManagerMock{}
	handlers.On("Connector", mock.Anything)
	handlers.On("Discoverer")

	// act
	go handleDialing(suite.log, &suite.wg, &suite.cfg, peers, pending, addresses, rep, handlers, stop)
	time.Sleep(time.Duration(1.5 * float64(suite.cfg.interval)))
	close(stop)
	suite.wg.Wait()

	// assert
	t := suite.T()

	handlers.AssertNotCalled(t, "Connector", mock.Anything)
	handlers.AssertNotCalled(t, "Discoverer")
}

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
	"errors"
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/awishformore/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestDropper(t *testing.T) {
	suite.Run(t, new(DropperSuite))
}

type DropperSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *DropperSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		interval: 2 * time.Millisecond,
		maxPeers: 15,
	}
}

func (suite *DropperSuite) TestDropperSuccess() {

	// arrange
	address := "192.0.2.100:1337"
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(16)
	peers.On("Addresses").Return([]string{address})
	peers.On("Drop", address).Return(nil)

	// act
	go handleDropping(suite.log, &suite.wg, &suite.cfg, peers, stop)
	time.Sleep(time.Duration(1.5 * float64(suite.cfg.interval)))
	close(stop)
	suite.wg.Wait()

	// assert
	t := suite.T()

	peers.AssertCalled(t, "Drop", address)
}

func (suite *DropperSuite) TestDropperValidPeerNumber() {

	// arrange
	address := "192.0.2.100:1337"
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(5)
	peers.On("Addresses").Return([]string{address})
	peers.On("Drop", address).Return(nil)

	// act
	go handleDropping(suite.log, &suite.wg, &suite.cfg, peers, stop)
	time.Sleep(time.Duration(1.5 * float64(suite.cfg.interval)))
	close(stop)
	suite.wg.Wait()

	// assert
	t := suite.T()

	peers.AssertNotCalled(t, "Drop", mock.Anything)
}

func (suite *DropperSuite) TestDropperDropFails() {

	// arrange
	address := "192.0.2.100:1337"
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(16)
	peers.On("Addresses").Return([]string{address})
	peers.On("Drop", address).Return(errors.New("could not drop peer"))

	// act
	go handleDropping(suite.log, &suite.wg, &suite.cfg, peers, stop)
	time.Sleep(time.Duration(2.5 * float64(suite.cfg.interval)))
	close(stop)
	suite.wg.Wait()

	// assert
	t := suite.T()

	peers.AssertCalled(t, "Drop", address)
	peers.AssertNumberOfCalls(t, "Drop", 2)
}

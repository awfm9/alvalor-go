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

type DialerTestSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *DialerTestSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		interval: 5 * time.Millisecond,
		minPeers: 5,
		maxPeers: 15,
	}
}

func (suite *DialerTestSuite) TestHandleDialingDoesNotDialIfMinAmountOfPeersConnected() {
	// arrange
	infos := &dialerInfosMock{}
	actions := &dialerActionsMock{}
	stop := make(chan struct{})

	actions.On("StartConnector")
	infos.On("PeerCount").Return(uint(5))
	infos.On("PendingCount").Return(uint(3))

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	handleDialing(suite.log, &suite.wg, &suite.cfg, infos, actions, stop)

	// assert
	actions.AssertNotCalled(suite.T(), "StartConnector")
}

func (suite *DialerTestSuite) TestHandleDialingDoesNotDialIfPendingAlreadyConnecting() {
	// arrange
	infos := &dialerInfosMock{}
	actions := &dialerActionsMock{}
	stop := make(chan struct{})

	actions.On("StartConnector")
	infos.On("PeerCount").Return(uint(3))
	infos.On("PendingCount").Return(uint(12))

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	handleDialing(suite.log, &suite.wg, &suite.cfg, infos, actions, stop)

	// assert
	actions.AssertNotCalled(suite.T(), "StartConnector")
}

func (suite *DialerTestSuite) TestHandleDialingDialsIfMinAmountOfPeersNotConnected() {
	// arrange
	infos := &dialerInfosMock{}
	actions := &dialerActionsMock{}
	stop := make(chan struct{})

	actions.On("StartConnector")
	infos.On("PeerCount").Return(uint(3))
	infos.On("PendingCount").Return(uint(5))

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	handleDialing(suite.log, &suite.wg, &suite.cfg, infos, actions, stop)

	// assert
	actions.AssertCalled(suite.T(), "StartConnector")
}

func TestDialerTestSuite(t *testing.T) {
	suite.Run(t, new(DialerTestSuite))
}

type dialerInfosMock struct {
	mock.Mock
}

func (infos *dialerInfosMock) PeerCount() uint {
	args := infos.Called()
	return args.Get(0).(uint)
}

func (infos *dialerInfosMock) PendingCount() uint {
	args := infos.Called()
	return args.Get(0).(uint)
}

type dialerActionsMock struct {
	mock.Mock
}

func (actions *dialerActionsMock) StartConnector() {
	actions.Called()
}

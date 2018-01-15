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

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type DropperTestSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *DropperTestSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		interval: 5 * time.Millisecond,
		maxPeers: 15,
	}
}

func (suite *DropperTestSuite) TestHandleDroppingDoesNotDropIfPeerCountLessThanMaxPeers() {
	// arrange
	infos := &dropperInfosMock{}
	actions := &dropperActionsMock{}
	events := &dropperEventsMock{}
	stop := make(chan struct{})

	addr := "33.22.72.33:525"
	actions.On("DropRandomPeer").Return(addr, nil)
	infos.On("PeerCount").Return(uint(5))
	events.On("Dropped", addr)

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	handleDropping(suite.log, &suite.wg, &suite.cfg, infos, actions, events, stop)

	// assert
	actions.AssertNotCalled(suite.T(), "DropRandomPeer")
}

func (suite *DropperTestSuite) TestHandleDroppingDropsConnectionIfPeerCountGreaterThanMaxPeers() {
	// arrange
	infos := &dropperInfosMock{}
	actions := &dropperActionsMock{}
	events := &dropperEventsMock{}
	stop := make(chan struct{})

	addr := "33.22.72.33:525"
	actions.On("DropRandomPeer").Return(addr, nil)
	infos.On("PeerCount").Return(uint(16))
	events.On("Dropped", addr)

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	handleDropping(suite.log, &suite.wg, &suite.cfg, infos, actions, events, stop)

	// assert
	actions.AssertCalled(suite.T(), "DropRandomPeer")
	events.AssertCalled(suite.T(), "Dropped", addr)
}

func (suite *DropperTestSuite) TestHandleDroppingDoesNotPublishEventIfDropNotSuccessful() {
	// arrange
	infos := &dropperInfosMock{}
	actions := &dropperActionsMock{}
	events := &dropperEventsMock{}
	stop := make(chan struct{})

	addr := ""
	actions.On("DropRandomPeer").Return(addr, errors.New("Can't drop any peer"))
	infos.On("PeerCount").Return(uint(16))
	events.On("Dropped", addr)

	go func() {
		time.Sleep(50 * time.Millisecond)
		stop <- struct{}{}
	}()

	// act
	handleDropping(suite.log, &suite.wg, &suite.cfg, infos, actions, events, stop)

	// assert
	events.AssertNotCalled(suite.T(), "Dropped", addr)
}

func TestDropperTestSuite(t *testing.T) {
	suite.Run(t, new(DropperTestSuite))
}

type dropperInfosMock struct {
	mock.Mock
}

func (infos *dropperInfosMock) PeerCount() uint {
	args := infos.Called()
	return args.Get(0).(uint)
}

type dropperActionsMock struct {
	mock.Mock
}

func (actions *dropperActionsMock) DropRandomPeer() (string, error) {
	args := actions.Called()
	return args.String(0), args.Error(1)
}

type dropperEventsMock struct {
	mock.Mock
}

func (events *dropperEventsMock) Dropped(addr string) {
	events.Called(addr)
}

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

	"github.com/awishformore/zerolog"
	"github.com/stretchr/testify/suite"
)

func TestServer(t *testing.T) {
	suite.Run(t, new(Server))
}

type Server struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *Server) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		interval: 2 * time.Millisecond,
		maxPeers: 5,
	}
}

func (suite *Server) TestServerSuccess() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(4)

	handlers := &HandlerManagerMock{}
	handlers.On("Listener")

	// act
	suite.cfg.listen = true
	go handleServing(suite.log, &suite.wg, &suite.cfg, peers, handlers, stop)
	time.Sleep(time.Duration(1.5 * float64(suite.cfg.interval)))
	close(stop)
	suite.wg.Wait()

	// assert
	handlers.AssertNumberOfCalls(suite.T(), "Listener", 1)
}

func (suite *Server) TestServerMaxPeersNotRunning() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(5)

	handlers := &HandlerManagerMock{}
	handlers.On("Listener")

	// act
	suite.cfg.listen = true
	go handleServing(suite.log, &suite.wg, &suite.cfg, peers, handlers, stop)
	time.Sleep(time.Duration(1.5 * float64(suite.cfg.interval)))
	close(stop)
	suite.wg.Wait()

	// assert
	handlers.AssertNotCalled(suite.T(), "Listener")
}

func (suite *Server) TestServerNotListening() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(4)

	handlers := &HandlerManagerMock{}
	handlers.On("Listener")

	// act
	suite.cfg.listen = false
	go handleServing(suite.log, &suite.wg, &suite.cfg, peers, handlers, stop)
	time.Sleep(time.Duration(1.5 * float64(suite.cfg.interval)))
	close(stop)
	suite.wg.Wait()

	// assert
	handlers.AssertNotCalled(suite.T(), "Listener")
}

func (suite *Server) TestServerMaxPeersRunning() {

	// arrange
	stop := make(chan struct{})

	peers := &PeerManagerMock{}
	peers.On("Count").Return(4).Once()
	peers.On("Count").Return(5)

	handlers := &HandlerManagerMock{}
	handlers.On("Listener")

	// act
	suite.cfg.listen = true
	go handleServing(suite.log, &suite.wg, &suite.cfg, peers, handlers, stop)
	time.Sleep(time.Duration(3.5 * float64(suite.cfg.interval)))
	close(stop)
	suite.wg.Wait()

	// assert
	handlers.AssertNumberOfCalls(suite.T(), "Listener", 1)
}

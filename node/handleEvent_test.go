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

package node

import (
	"io/ioutil"
	"sync"
	"testing"

	"github.com/awishformore/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/alvalor/alvalor-go/network"
)

func TestEvent(t *testing.T) {
	suite.Run(t, new(EventSuite))
}

type EventSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  *sync.WaitGroup
}

func (suite *EventSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = &sync.WaitGroup{}
	suite.wg.Add(1)
}

func (suite *EventSuite) TestEventConnected() {

	// arrange
	address := "192.0.2.1:1337"

	net := &NetworkMock{}
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	chain := &BlockchainMock{}

	peers := &PeersMock{}
	peers.On("Active", mock.Anything)
	peers.On("Inactive", mock.Anything)
	peers.On("Tag", mock.Anything, mock.Anything)

	pool := &PoolMock{}
	pool.On("Count").Return(0)
	pool.On("Hashes").Return([][]byte{})

	handlers := &HandlersMock{}
	handlers.On("Message", mock.Anything)

	event := network.Connected{Address: address}

	// act
	handleEvent(suite.log, suite.wg, net, chain, peers, handlers, event)

	// assert
	t := suite.T()

	peers.AssertCalled(t, "Active", address)
}

func (suite *EventSuite) TestEventDisconnected() {

	// arrange
	address := "192.0.2.1:1337"

	net := &NetworkMock{}
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	chain := &BlockchainMock{}

	peers := &PeersMock{}
	peers.On("Active", mock.Anything)
	peers.On("Inactive", mock.Anything)
	peers.On("Tag", mock.Anything, mock.Anything)

	pool := &PoolMock{}
	pool.On("Count").Return(0)
	pool.On("IDs").Return([][]byte{})

	handlers := &HandlersMock{}
	handlers.On("Message", mock.Anything)

	event := network.Disconnected{Address: address}

	// act
	handleEvent(suite.log, suite.wg, net, chain, peers, handlers, event)

	// assert
	t := suite.T()

	peers.AssertCalled(t, "Inactive", address)
}

func (suite *EventSuite) TestEventReceived() {

	// arrange
	address := "192.0.2.1:1337"
	message := "message"

	net := &NetworkMock{}
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	chain := &BlockchainMock{}

	peers := &PeersMock{}
	peers.On("Active", mock.Anything)
	peers.On("Inactive", mock.Anything)
	peers.On("Tag", mock.Anything, mock.Anything)

	pool := &PoolMock{}
	pool.On("Count").Return(0)
	pool.On("IDs").Return([][]byte{})

	handlers := &HandlersMock{}
	handlers.On("Message", mock.Anything, mock.Anything)

	event := network.Received{Address: address, Message: message}

	// act
	handleEvent(suite.log, suite.wg, net, chain, peers, handlers, event)

	// assert
	t := suite.T()

	handlers.AssertCalled(t, "Message", address, message)
}

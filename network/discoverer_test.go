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
	"errors"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/awishformore/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestDiscoverer(t *testing.T) {
	suite.Run(t, new(DiscovererSuite))
}

type DiscovererSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *DiscovererSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{}
}

func (suite *ConnectorSuite) TestDiscovererSuccess() {

	// arrange
	address1 := "192.0.2.10:1337"
	address2 := "192.0.2.20:1337"
	address3 := "192.0.2.30:1337"

	peers := &PeerManagerMock{}
	peers.On("Addresses").Return([]string{address1, address2, address3})
	peers.On("Send", mock.Anything, mock.Anything).Return(nil)

	// act
	handleDiscovering(suite.log, &suite.wg, &suite.cfg, peers)

	// assert
	t := suite.T()

	peers.AssertCalled(t, "Send", address1, &Discover{})
	peers.AssertCalled(t, "Send", address2, &Discover{})
	peers.AssertCalled(t, "Send", address3, &Discover{})
}

func (suite *ConnectorSuite) TestDiscovererNoPeers() {

	// arrange
	peers := &PeerManagerMock{}
	peers.On("Addresses").Return([]string{})
	peers.On("Send", mock.Anything, mock.Anything).Return(nil)

	// act
	handleDiscovering(suite.log, &suite.wg, &suite.cfg, peers)

	// assert
	t := suite.T()

	peers.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
}

func (suite *ConnectorSuite) TestDiscovererSendFails() {

	// arrange
	address1 := "192.0.2.10:1337"
	address2 := "192.0.2.20:1337"
	address3 := "192.0.2.30:1337"

	peers := &PeerManagerMock{}
	peers.On("Addresses").Return([]string{address1, address2, address3})
	peers.On("Send", mock.Anything, mock.Anything).Return(nil).Once()
	peers.On("Send", mock.Anything, mock.Anything).Return(errors.New("could not send discover"))
	peers.On("Send", mock.Anything, mock.Anything).Return(nil)

	// act
	handleDiscovering(suite.log, &suite.wg, &suite.cfg, peers)

	// assert
	t := suite.T()

	peers.AssertCalled(t, "Send", address1, &Discover{})
	peers.AssertCalled(t, "Send", address2, &Discover{})
	peers.AssertCalled(t, "Send", address3, &Discover{})
}

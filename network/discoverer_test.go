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

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
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

func (suite *ConnectorSuite) TestDiscovererNoPeers() {

	// arrange
	closed := make(chan struct{})
	close(closed)

	peers := &PeerManagerMock{}
	peers.On("Addresses").Return(nil)
	peers.On("Output", mock.Anything).Return(closed, nil)

	// act
	handleDiscovering(suite.log, &suite.wg, &suite.cfg, peers)

	// assert
	peers.AssertCalled(suite.T(), "Addresses")
	peers.AssertNotCalled(suite.T(), "Output")
}

func (suite *ConnectorSuite) TestDiscovererMissingOutput() {

	// arrange
	address1 := "192.0.2.10:1337"
	address2 := "192.0.2.20:1337"
	address3 := "192.0.2.30:1337"

	output := make(chan interface{}, 3)

	peers := &PeerManagerMock{}
	peers.On("Addresses").Return([]string{address1, address2, address3})
	peers.On("Output", address1).Return(output, nil)
	peers.On("Output", address2).Return(nil, errors.New("could not get channel"))
	peers.On("Output", address3).Return(output, nil)

	// act
	handleDiscovering(suite.log, &suite.wg, &suite.cfg, peers)
	close(output)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}

	// assert
	peers.AssertCalled(suite.T(), "Addresses")
	peers.AssertNumberOfCalls(suite.T(), "Output", 3)
	if assert.Len(suite.T(), msgs, 2) {
		assert.IsType(suite.T(), &Discover{}, msgs[0])
		assert.IsType(suite.T(), &Discover{}, msgs[1])
	}
}

func (suite *ConnectorSuite) TestDiscovererSuccess() {

	// arrange
	address1 := "192.0.2.10:1337"
	address2 := "192.0.2.20:1337"
	address3 := "192.0.2.30:1337"

	output := make(chan interface{}, 3)

	peers := &PeerManagerMock{}
	peers.On("Addresses").Return([]string{address1, address2, address3})
	peers.On("Output", address1).Return(output, nil)
	peers.On("Output", address2).Return(output, nil)
	peers.On("Output", address3).Return(output, nil)

	// act
	handleDiscovering(suite.log, &suite.wg, &suite.cfg, peers)
	close(output)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}

	// assert
	peers.AssertCalled(suite.T(), "Addresses")
	peers.AssertNumberOfCalls(suite.T(), "Output", 3)
	if assert.Len(suite.T(), msgs, 3) {
		assert.IsType(suite.T(), &Discover{}, msgs[0])
		assert.IsType(suite.T(), &Discover{}, msgs[1])
		assert.IsType(suite.T(), &Discover{}, msgs[2])
	}
}

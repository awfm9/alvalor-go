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
	"errors"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/awishformore/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/alvalor/alvalor-go/types"
)

func TestEntity(t *testing.T) {
	suite.Run(t, new(EntitySuite))
}

type EntitySuite struct {
	suite.Suite
	log zerolog.Logger
	wg  *sync.WaitGroup
}

func (suite *EntitySuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = &sync.WaitGroup{}
	suite.wg.Add(1)
}

func (suite *EntitySuite) TestEntityTransaction() {

	// arrange
	address1 := "192.0.2.1:1337"
	address2 := "192.0.2.2:1337"
	address3 := "192.0.2.3:1337"

	net := &NetworkMock{}
	net.On("Send", address1, mock.Anything).Return(errors.New("could not send"))
	net.On("Send", address2, mock.Anything).Return(nil)
	net.On("Send", address3, mock.Anything).Return(nil)

	peers := &PeersMock{}
	peers.On("Tags", mock.Anything).Return([]string{address2})
	peers.On("Actives").Return([]string{address1, address2, address3})

	pool := &PoolMock{}
	pool.On("Known", mock.Anything).Return(false)
	pool.On("Add", mock.Anything).Return(nil)

	entity := &types.Transaction{}

	// act
	handleEntity(suite.log, suite.wg, net, peers, pool, entity)

	// assert
	t := suite.T()

	net.AssertCalled(t, "Send", address1, mock.Anything)
	net.AssertCalled(t, "Send", address3, mock.Anything)

	net.AssertNotCalled(t, "Send", address2, mock.Anything)
}

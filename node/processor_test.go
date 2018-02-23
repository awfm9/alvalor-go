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

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/alvalor/alvalor-go/network"
)

func TestProcessor(t *testing.T) {
	suite.Run(t, new(ProcessorSuite))
}

type ProcessorSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  *sync.WaitGroup
}

func (suite *ProcessorSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = &sync.WaitGroup{}
	suite.wg.Add(1)
}

func (suite *ProcessorSuite) TestProcessorTransactionNew() {

	// arrange
	pool := &PoolMock{}
	pool.On("Known", mock.Anything).Return(false)
	pool.On("Add", mock.Anything).Return(nil)

	handlers := &HandlersMock{}
	handlers.On("Propagate", mock.Anything)

	net := &NetworkMock{}

	event := network.Received{}

	// act
	handleProcessing(suite.log, suite.wg, handlers, pool, net, event)

	// assert
	t := suite.T()

	// pool.AssertCalled(t, "Known", tx.ID())
	// pool.AssertCalled(t, "Add", tx)
	handlers.AssertCalled(t, "Propagate", event.Message)
}

func (suite *ProcessorSuite) TestProcessorTransactionKnown() {

	// arrange
	pool := &PoolMock{}
	pool.On("Known", mock.Anything).Return(true)
	pool.On("Add", mock.Anything).Return(nil)

	handlers := &HandlersMock{}
	handlers.On("Propagate", mock.Anything)

	net := &NetworkMock{}

	event := network.Received{}

	// act
	handleProcessing(suite.log, suite.wg, handlers, pool, net, event)
	// assert
	t := suite.T()

	// pool.AssertCalled(t, "Known", tx.ID())

	pool.AssertNotCalled(t, "Add", mock.Anything)
	handlers.AssertNotCalled(t, "Propagate", mock.Anything)
}

func (suite *ProcessorSuite) TestProcessorTransactionAddFails() {

	// arrange
	pool := &PoolMock{}
	pool.On("Known", mock.Anything).Return(false)
	pool.On("Add", mock.Anything).Return(errors.New("could not add"))

	handlers := &HandlersMock{}
	handlers.On("Propagate", mock.Anything)

	net := &NetworkMock{}

	event := network.Received{}

	// act
	handleProcessing(suite.log, suite.wg, handlers, pool, net, event)

	// assert
	t := suite.T()

	// pool.AssertCalled(t, "Known", tx.ID())
	// pool.AssertCalled(t, "Add", tx)

	handlers.AssertNotCalled(t, "Propagate", mock.Anything)
}

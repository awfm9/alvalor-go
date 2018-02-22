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

	"github.com/alvalor/alvalor-go/network"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestReceiver(t *testing.T) {
	suite.Run(t, new(ReceiverSuite))
}

type ReceiverSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  *sync.WaitGroup
}

func (suite *ReceiverSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = &sync.WaitGroup{}
	suite.wg.Add(1)
}

func (suite *ReceiverSuite) TestReceiverConnected() {

	// arrange
	address := "192.0.2.1:1337"

	sub := make(chan interface{})

	handlers := &HandlersMock{}
	handlers.On("Process", mock.Anything)

	state := &StateMock{}
	state.On("Active", mock.Anything)
	state.On("Inactive", mock.Anything)
	state.On("Tag", mock.Anything, mock.Anything)

	// act
	go handleReceiving(suite.log, suite.wg, sub, handlers, state)
	sub <- &network.Connected{Address: address}
	close(sub)
	suite.wg.Wait()

	// assert
	t := suite.T()

	state.AssertCalled(t, "Active", address)
}

func (suite *ReceiverSuite) TestReceiverDisconnected() {

	// arrange
	address := "192.0.2.1:1337"

	sub := make(chan interface{})

	handlers := &HandlersMock{}
	handlers.On("Process", mock.Anything)

	state := &StateMock{}
	state.On("Active", mock.Anything)
	state.On("Inactive", mock.Anything)
	state.On("Tag", mock.Anything, mock.Anything)

	// act
	go handleReceiving(suite.log, suite.wg, sub, handlers, state)
	sub <- &network.Disconnected{Address: address}
	close(sub)
	suite.wg.Wait()

	// assert
	t := suite.T()

	state.AssertCalled(t, "Inactive", address)
}

func (suite *ReceiverSuite) TestReceiverReceived() {

	// arrange
	address := "192.0.2.1:1337"
	id := []byte{1, 2, 3, 4}

	entity := &EntityMock{}
	entity.On("ID").Return(id)

	handlers := &HandlersMock{}
	handlers.On("Process", mock.Anything)

	state := &StateMock{}
	state.On("Active", mock.Anything)
	state.On("Inactive", mock.Anything)
	state.On("Tag", mock.Anything, mock.Anything)

	sub := make(chan interface{})

	// act
	go handleReceiving(suite.log, suite.wg, sub, handlers, state)
	sub <- &network.Received{Address: address, Message: entity}
	close(sub)
	suite.wg.Wait()

	// assert
	t := suite.T()

	state.AssertCalled(t, "Tag", address, id)
	handlers.AssertCalled(t, "Process", entity)
}

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestProcessor(t *testing.T) {
	suite.Run(t, new(ProcessorSuite))
}

type ProcessorSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *ProcessorSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		interval: 2 * time.Millisecond,
	}
}

func (suite *ProcessorSuite) TestProcessorSuccess() {

	// arrange
	address := "192.0.2.100:1337"
	sample := []string{"192.0.2.200:1337", "192.0.2.201:1337", "192.0.2.202:1337"}

	input := make(chan interface{})
	output := make(chan interface{}, 5)

	book := &AddressManagerMock{}
	book.On("Add", mock.Anything)
	book.On("Sample", mock.Anything, mock.Anything).Return(sample)

	events := &EventManagerMock{}
	events.On("Received", mock.Anything, mock.Anything).Return(nil)

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, book, events, address, input, output)
	close(input)
	suite.wg.Wait()
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}

	// assert
	t := suite.T()

	if assert.Len(t, msgs, 1) {
		assert.IsType(t, &Discover{}, msgs[0])
	}
}

func (suite *ProcessorSuite) TestProcessorTimeout() {

	// arrange
	address := "192.0.2.100:1337"
	sample := []string{"192.0.2.200:1337", "192.0.2.201:1337", "192.0.2.202:1337"}

	input := make(chan interface{})
	output := make(chan interface{}, 5)

	book := &AddressManagerMock{}
	book.On("Add", mock.Anything)
	book.On("Sample", mock.Anything, mock.Anything).Return(sample)

	events := &EventManagerMock{}
	events.On("Received", mock.Anything, mock.Anything).Return(nil)

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, book, events, address, input, output)
	time.Sleep(time.Duration(4.5 * float64(suite.cfg.interval)))
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
}

func (suite *ProcessorSuite) TestProcessorUnknownMessage() {

	// arrange
	address := "192.0.2.100:1337"
	sample := []string{"192.0.2.200:1337", "192.0.2.201:1337", "192.0.2.202:1337"}

	input := make(chan interface{})
	output := make(chan interface{}, 5)

	messages := []interface{}{
		1337,
		"message",
		[]byte{1, 2, 3, 4, 5},
		map[string]bool{"field": true},
		&struct{ field int }{field: 7},
		true, // discarded
		true, // discarded
		true, // discarded
	}

	book := &AddressManagerMock{}
	book.On("Add", mock.Anything)
	book.On("Sample", mock.Anything, mock.Anything).Return(sample)

	events := &EventManagerMock{}
	events.On("Received", mock.Anything, mock.Anything).Return(nil)

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, book, events, address, input, output)
	for _, msg := range messages {
		input <- msg
	}
	close(input)
	suite.wg.Wait()

	// assert
	t := suite.T()

	events.AssertCalled(t, "Received", address, messages[0])
	events.AssertCalled(t, "Received", address, messages[1])
	events.AssertCalled(t, "Received", address, messages[2])
	events.AssertCalled(t, "Received", address, messages[3])
	events.AssertCalled(t, "Received", address, messages[4])
}

func (suite *ProcessorSuite) TestProcessorPing() {

	// arrange
	address := "192.0.2.100:1337"
	sample := []string{"192.0.2.200:1337", "192.0.2.201:1337", "192.0.2.202:1337"}

	input := make(chan interface{})
	output := make(chan interface{}, 5)

	book := &AddressManagerMock{}
	book.On("Add", mock.Anything)
	book.On("Sample", mock.Anything, mock.Anything).Return(sample)

	events := &EventManagerMock{}
	events.On("Received", mock.Anything, mock.Anything).Return(nil)

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, book, events, address, input, output)
	input <- &Ping{}
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	t := suite.T()

	if assert.Len(t, msgs, 2) {
		assert.IsType(t, &Pong{}, msgs[1])
	}
}

func (suite *ProcessorSuite) TestProcessorDiscover() {

	// arrange
	address := "192.0.2.100:1337"
	sample := []string{"192.0.2.200:1337", "192.0.2.201:1337", "192.0.2.202:1337"}

	input := make(chan interface{})
	output := make(chan interface{}, 5)

	book := &AddressManagerMock{}
	book.On("Add", mock.Anything)
	book.On("Sample", mock.Anything, mock.Anything).Return(sample)

	events := &EventManagerMock{}
	events.On("Received", mock.Anything, mock.Anything).Return(nil)

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, book, events, address, input, output)
	input <- &Discover{}
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	t := suite.T()

	if assert.Len(t, msgs, 2) {
		assert.IsType(t, &Peers{}, msgs[1])
		peersMsg := msgs[1].(*Peers)
		assert.Equal(t, sample, peersMsg.Addresses)
	}
}

func (suite *ProcessorSuite) TestProcessorPeers() {

	// arrange
	address := "192.0.2.100:1337"
	sample := []string{"192.0.2.200:1337", "192.0.2.201:1337", "192.0.2.202:1337"}

	peer1 := "192.0.2.250:1337"
	peer2 := "192.0.2.251:1337"
	peer3 := "192.0.2.252:1337"

	input := make(chan interface{})
	output := make(chan interface{}, 5)

	book := &AddressManagerMock{}
	book.On("Add", mock.Anything)
	book.On("Sample", mock.Anything, mock.Anything).Return(sample)

	events := &EventManagerMock{}
	events.On("Received", mock.Anything, mock.Anything).Return(nil)

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, book, events, address, input, output)
	input <- &Peers{Addresses: []string{peer1, peer2, peer3}}
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	t := suite.T()

	if book.AssertNumberOfCalls(t, "Add", 3) {
		book.AssertCalled(t, "Add", peer1)
		book.AssertCalled(t, "Add", peer2)
		book.AssertCalled(t, "Add", peer3)
	}
}

func (suite *ProcessorSuite) TestProcessorPong() {

	// arrange
	address := "192.0.2.100:1337"
	sample := []string{"192.0.2.200:1337", "192.0.2.201:1337", "192.0.2.202:1337"}

	input := make(chan interface{})
	output := make(chan interface{}, 5)

	book := &AddressManagerMock{}
	book.On("Add", mock.Anything)
	book.On("Sample", mock.Anything, mock.Anything).Return(sample)

	events := &EventManagerMock{}
	events.On("Received", mock.Anything, mock.Anything).Return(nil)

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, book, events, address, input, output)
	input <- &Pong{}
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
}

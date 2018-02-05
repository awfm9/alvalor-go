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

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
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
		interval: 10 * time.Millisecond,
		address:  "192.0.2.100:1337",
		listen:   false,
	}
}

func (suite *ProcessorSuite) TestProcessingEnabledListenPublishesOwnAddress() {

	// arrange
	address := "192.0.2.200:1337"
	input := make(chan interface{})
	output := make(chan interface{}, 16)

	addresses := &AddressManagerMock{}

	peers := &PeerManagerMock{}

	// act
	suite.cfg.listen = true
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, addresses, peers, nil, address, input, output)
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	if assert.Len(suite.T(), msgs, 2) {
		assert.IsType(suite.T(), &Peers{}, msgs[0])
		assert.IsType(suite.T(), &Discover{}, msgs[1])
		peersMsg := msgs[0].(*Peers)
		assert.Equal(suite.T(), []string{suite.cfg.address}, peersMsg.Addresses)
	}
}

func (suite *ProcessorSuite) TestProcessingPublishesDiscoverNotOwnAddress() {

	// arrange
	address := "192.0.2.200:1337"
	input := make(chan interface{})
	output := make(chan interface{}, 16)

	addresses := &AddressManagerMock{}

	peers := &PeerManagerMock{}

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, addresses, peers, nil, address, input, output)
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	if assert.Len(suite.T(), msgs, 1) {
		assert.IsType(suite.T(), &Discover{}, msgs[0])
	}
}

func (suite *ProcessorSuite) TestProcessingSendsPingEachInterval() {

	// arrange
	address := "192.0.2.200:1337"
	input := make(chan interface{})
	output := make(chan interface{}, 16)

	addresses := &AddressManagerMock{}

	peers := &PeerManagerMock{}

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, addresses, peers, nil, address, input, output)
	time.Sleep(time.Duration(2.5 * float64(suite.cfg.interval)))
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	if assert.Len(suite.T(), msgs, 3) {
		assert.IsType(suite.T(), &Ping{}, msgs[1])
		assert.IsType(suite.T(), &Ping{}, msgs[2])
	}
}

func (suite *ProcessorSuite) TestProcessingRespondsToPingWithPong() {

	// arrange
	address := "192.0.2.200:1337"
	input := make(chan interface{})
	output := make(chan interface{}, 16)

	addresses := &AddressManagerMock{}

	peers := &PeerManagerMock{}

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, addresses, peers, nil, address, input, output)
	input <- &Ping{}
	input <- &Ping{}
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	if assert.Len(suite.T(), msgs, 3) {
		assert.IsType(suite.T(), &Pong{}, msgs[1])
		assert.IsType(suite.T(), &Pong{}, msgs[2])
	}
}

func (suite *ProcessorSuite) TestProcessingRespondsToDiscoverWithPeers() {

	// arrange
	address := "192.0.2.200:1337"
	sample := []string{"15.77.14.74:6666", "15.77.14.74:7777", "15.77.14.74:8888"}
	input := make(chan interface{})
	output := make(chan interface{}, 16)

	addresses := &AddressManagerMock{}
	addresses.On("Sample", 8).Return(sample)

	peers := &PeerManagerMock{}

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, addresses, peers, nil, address, input, output)
	input <- &Discover{}
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	if assert.Len(suite.T(), msgs, 2) {
		assert.IsType(suite.T(), &Peers{}, msgs[1])
		peersMsg := msgs[1].(*Peers)
		assert.Equal(suite.T(), sample, peersMsg.Addresses)
	}
}

func (suite *ProcessorSuite) TestProcessingAddsPeersAddresses() {

	// arrange
	address := "192.0.2.200:1337"
	sample := []string{"15.77.14.74:6666", "15.77.14.74:7777", "15.77.14.74:8888"}
	input := make(chan interface{})
	output := make(chan interface{}, 16)

	addresses := &AddressManagerMock{}
	addresses.On("Add", sample[0])
	addresses.On("Add", sample[1])
	addresses.On("Add", sample[2])

	peers := &PeerManagerMock{}

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, addresses, peers, nil, address, input, output)
	input <- &Peers{Addresses: sample}
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	if addresses.AssertNumberOfCalls(suite.T(), "Add", 3) {
		addresses.AssertCalled(suite.T(), "Add", sample[0])
		addresses.AssertCalled(suite.T(), "Add", sample[1])
		addresses.AssertCalled(suite.T(), "Add", sample[2])
	}
}

func (suite *ProcessorSuite) TestProcessingPong() {

	// arrange
	address := "192.0.2.200:1337"
	input := make(chan interface{})
	output := make(chan interface{}, 16)

	addresses := &AddressManagerMock{}

	peers := &PeerManagerMock{}

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, addresses, peers, nil, address, input, output)
	input <- &Pong{}
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
}

func (suite *ProcessorSuite) TestProcessingDropsPeerAfterThreePings() {

	// arrange
	address := "192.0.2.200:1337"
	input := make(chan interface{})
	output := make(chan interface{}, 16)

	addresses := &AddressManagerMock{}

	peers := &PeerManagerMock{}
	peers.On("Drop", address).Return(nil)

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, addresses, peers, nil, address, input, output)
	time.Sleep(time.Duration(3.75 * float64(suite.cfg.interval)))
	close(input)
	var msgs []interface{}
	for msg := range output {
		msgs = append(msgs, msg)
	}
	suite.wg.Wait()

	// assert
	if assert.Len(suite.T(), msgs, 4) {
		assert.IsType(suite.T(), &Ping{}, msgs[1])
		assert.IsType(suite.T(), &Ping{}, msgs[2])
		assert.IsType(suite.T(), &Ping{}, msgs[3])
		peers.AssertCalled(suite.T(), "Drop", address)
	}
}

func (suite *ProcessorSuite) TestProcessingForwardsCustomMessages() {

	// arrange
	address := "192.0.2.200:1337"
	input := make(chan interface{})
	output := make(chan interface{}, 16)
	subscriber := make(chan interface{}, 5)

	addresses := &AddressManagerMock{}

	peers := &PeerManagerMock{}

	// act
	go handleProcessing(suite.log, &suite.wg, &suite.cfg, addresses, peers, subscriber, address, input, output)
	messages := []interface{}{
		1337,
		"some message",
		[]byte{1, 2, 3, 4, 5},
		map[string]bool{"field": true},
		&struct{ field int }{field: 7},
		true, // discarded
		true, // discarded
		true, // discarded
	}
	for _, message := range messages {
		input <- message
	}
	close(input)
	suite.wg.Wait()
	var receiveds []interface{}
Loop:
	for {
		select {
		case received := <-subscriber:
			receiveds = append(receiveds, received)
		default:
			break Loop
		}
	}

	// assert
	if assert.Len(suite.T(), receiveds, 5) {
		msgs := make([]interface{}, 0, 5)
		for _, rcvd := range receiveds {
			assert.IsType(suite.T(), &Received{}, rcvd)
			received := rcvd.(*Received)
			msgs = append(msgs, received.Message)
		}
		assert.Equal(suite.T(), messages[:5], msgs)
	}
}

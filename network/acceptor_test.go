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
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestAcceptor(t *testing.T) {
	suite.Run(t, new(AcceptorSuite))
}

type AcceptorSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *AcceptorSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		network: Odin,
		nonce:   uuid.NewV4().Bytes(),
	}
}

func (suite *AcceptorSuite) TestAcceptorSuccess() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, nonce...)

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(0, nil)
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	peers := &PeerManagerMock{}
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)
	rep.On("Success", mock.Anything)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	events := &EventManagerMock{}
	events.On("Connected", mock.Anything).Return(nil)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, events, conn)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	peers.AssertCalled(t, "Add", conn, nonce)
	rep.AssertCalled(t, "Success", address)
	events.AssertCalled(t, "Connected", address)

	conn.AssertNotCalled(t, "Close")
	rep.AssertNotCalled(t, "Failure", mock.Anything)
	book.AssertNotCalled(t, "Block", mock.Anything)
}

func (suite *AcceptorSuite) TestAcceptorClaimFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, nonce...)

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(0, nil)
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(errors.New("could not claim slot"))
	pending.On("Release", mock.Anything).Return(nil)

	peers := &PeerManagerMock{}
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)
	rep.On("Success", mock.Anything)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	events := &EventManagerMock{}
	events.On("Connected", mock.Anything).Return(nil)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, events, conn)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	conn.AssertCalled(t, "Close")

	pending.AssertNotCalled(t, "Release", mock.Anything)
	peers.AssertNotCalled(t, "Add", mock.Anything, mock.Anything)
	rep.AssertNotCalled(t, "Success", mock.Anything)
	rep.AssertNotCalled(t, "Failure", mock.Anything)
	book.AssertNotCalled(t, "Block", mock.Anything)
	events.AssertNotCalled(t, "Connected", mock.Anything)
}

func (suite *AcceptorSuite) TestAcceptorReadFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, nonce...)

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(0, errors.New("could not read syn"))
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	peers := &PeerManagerMock{}
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)
	rep.On("Success", mock.Anything)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	events := &EventManagerMock{}
	events.On("Connected", mock.Anything).Return(nil)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, events, conn)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	rep.AssertCalled(t, "Failure", address)
	conn.AssertCalled(t, "Close")

	peers.AssertNotCalled(t, "Add", mock.Anything, mock.Anything)
	rep.AssertNotCalled(t, "Success", mock.Anything)
	book.AssertNotCalled(t, "Block", mock.Anything)
	events.AssertNotCalled(t, "Connected", mock.Anything)
}

func (suite *AcceptorSuite) TestAcceptorNetworkMismatch() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append([]byte{1, 2, 3, 4}, nonce...)

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(0, nil)
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	peers := &PeerManagerMock{}
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)
	rep.On("Success", mock.Anything)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	events := &EventManagerMock{}
	events.On("Connected", mock.Anything).Return(nil)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, events, conn)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	book.AssertCalled(t, "Block", address)
	conn.AssertCalled(t, "Close")

	peers.AssertNotCalled(t, "Add", mock.Anything, mock.Anything)
	rep.AssertNotCalled(t, "Success", mock.Anything)
	rep.AssertNotCalled(t, "Failure", mock.Anything)
	events.AssertNotCalled(t, "Connected", mock.Anything)
}

func (suite *AcceptorSuite) TestAcceptorNonceIdentical() {

	// arrange
	address := "192.0.2.100:1337"
	syn := append(suite.cfg.network, suite.cfg.nonce...)

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(0, nil)
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	peers := &PeerManagerMock{}
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)
	rep.On("Success", mock.Anything)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	events := &EventManagerMock{}
	events.On("Connected", mock.Anything)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, events, conn)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	book.AssertCalled(t, "Block", address)
	conn.AssertCalled(t, "Close")

	peers.AssertNotCalled(t, "Add", mock.Anything, mock.Anything)
	rep.AssertNotCalled(t, "Success", mock.Anything)
	rep.AssertNotCalled(t, "Failure", mock.Anything)
	events.AssertNotCalled(t, "Connected", mock.Anything)
}

func (suite *AcceptorSuite) TestAcceptorWriteFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, nonce...)

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(0, nil)
	conn.On("Write", mock.Anything).Return(0, errors.New("could not write ack"))
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	peers := &PeerManagerMock{}
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)
	rep.On("Success", mock.Anything)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	events := &EventManagerMock{}
	events.On("Connected", mock.Anything).Return(nil)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, events, conn)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	rep.AssertCalled(t, "Failure", address)
	conn.AssertCalled(t, "Close")

	peers.AssertNotCalled(t, "Add", mock.Anything, mock.Anything)
	rep.AssertNotCalled(t, "Success", mock.Anything)
	book.AssertNotCalled(t, "Block", mock.Anything)
	events.AssertNotCalled(t, "Connected", mock.Anything)
}

func (suite *AcceptorSuite) TestAcceptorAddPeerFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, nonce...)

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(0, nil)
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	peers := &PeerManagerMock{}
	peers.On("Add", mock.Anything, mock.Anything).Return(errors.New("could not add peer"))

	rep := &ReputationManagerMock{}
	rep.On("Failure", mock.Anything)
	rep.On("Success", mock.Anything)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	events := &EventManagerMock{}
	events.On("Connected", mock.Anything).Return(nil)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, events, conn)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	peers.AssertCalled(t, "Add", conn, nonce)
	conn.AssertCalled(t, "Close")

	rep.AssertNotCalled(t, "Success", mock.Anything)
	rep.AssertNotCalled(t, "Failure", mock.Anything)
	book.AssertNotCalled(t, "Block", mock.Anything)
	events.AssertNotCalled(t, "Connected", mock.Anything)
}

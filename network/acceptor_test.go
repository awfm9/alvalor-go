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

func (suite *AcceptorSuite) TestAcceptorClaimFails() {

	// arrange
	address := "192.0.2.100:1337"

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(errors.New("cannot claim slot"))

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}

	book := &AddressManagerMock{}

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, conn)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertNotCalled(suite.T(), "Release", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorSuite) TestAcceptorReadFails() {

	// arrange
	address := "192.0.2.100:1337"
	buf := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", buf).Return(1, errors.New("cannot read from connection"))
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}
	rep.On("Error", address)

	book := &AddressManagerMock{}

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, conn)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	rep.AssertCalled(suite.T(), "Error", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorSuite) TestAcceptorNetworkMismatch() {

	// arrange
	address := "192.0.2.100:1337"
	syn := append([]byte{1, 2, 3, 4}, uuid.NewV4().Bytes()...)
	buf := make([]byte, len(syn))

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}
	rep.On("Invalid", address)

	book := &AddressManagerMock{}

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, conn)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	rep.AssertCalled(suite.T(), "Invalid", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorSuite) TestAcceptorNonceIdentical() {

	// arrange
	address := "192.0.2.100:1337"
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}
	rep.On("Invalid", address)

	book := &AddressManagerMock{}

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, conn)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	rep.AssertCalled(suite.T(), "Invalid", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorSuite) TestAcceptorWriteFails() {

	// arrange
	address := "192.0.2.100:1337"
	syn := append(suite.cfg.network, uuid.NewV4().Bytes()...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, suite.cfg.nonce...)

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(len(buf), nil)
	conn.On("Write", ack).Return(0, errors.New("cannot write to connection"))
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}
	rep.On("Error", address)

	book := &AddressManagerMock{}

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, conn)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	rep.AssertCalled(suite.T(), "Error", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorSuite) TestAcceptorAddPeerFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, suite.cfg.nonce...)

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	peers := &PeerManagerMock{}
	peers.On("Add", conn, nonce).Return(errors.New("cannot add peer"))

	rep := &ReputationManagerMock{}
	rep.On("Error", address)

	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(len(buf), nil)
	conn.On("Write", ack).Return(len(ack), nil)
	conn.On("Close").Return(nil)

	book := &AddressManagerMock{}

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, conn)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorSuite) TestAcceptorSuccess() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, suite.cfg.nonce...)

	addr := &AddrMock{}
	addr.On("String").Return(address)

	conn := &ConnMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(len(buf), nil)
	conn.On("Write", ack).Return(len(ack), nil)
	conn.On("Close").Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	peers := &PeerManagerMock{}
	peers.On("Add", conn, nonce).Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", address)

	book := &AddressManagerMock{}

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, conn)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	peers.AssertCalled(suite.T(), "Add", conn, nonce)
	rep.AssertCalled(suite.T(), "Success", address)
}

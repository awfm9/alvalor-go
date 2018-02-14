// // Copyright (c) 2017 The Alvalor Authors
// //
// // This file is part of Alvalor.
// //
// // Alvalor is free software: you can redistribute it and/or modify
// // it under the terms of the GNU Affero General Public License as published by
// // the Free Software Foundation, either version 3 of the License, or
// // (at your option) any later version.
// //
// // Alvalor is distributed in the hope that it will be useful,
// // but WITHOUT ANY WARRANTY; without even the implied warranty of
// // MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// // GNU Affero General Public License for more details.
// //
// // You should have received a copy of the GNU Affero General Public License
// // along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.
//
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

func TestConnector(t *testing.T) {
	suite.Run(t, new(ConnectorSuite))
}

type ConnectorSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *ConnectorSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		network: Odin,
		nonce:   uuid.NewV4().Bytes(),
	}
}

func (suite *ConnectorSuite) TestConnectorSuccess() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	ack := append(suite.cfg.network, nonce...)

	conn := &ConnMock{}
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(0, nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", mock.Anything)
	rep.On("Failure", mock.Anything)

	peers := &PeerManagerMock{}
	peers.On("Known", mock.Anything).Return(false)
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	dialer := &DialManagerMock{}
	dialer.On("Dial", mock.Anything).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	peers.AssertCalled(t, "Add", conn, nonce)
	rep.AssertCalled(t, "Success", address)

	conn.AssertNotCalled(t, "Close")
	rep.AssertNotCalled(t, "Failure")
	book.AssertNotCalled(t, "Block")
}

func (suite *ConnectorSuite) TestConnectorClaimFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	ack := append(suite.cfg.network, nonce...)

	conn := &ConnMock{}
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(0, nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", mock.Anything)
	rep.On("Failure", mock.Anything)

	peers := &PeerManagerMock{}
	peers.On("Known", mock.Anything).Return(false)
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	dialer := &DialManagerMock{}
	dialer.On("Dial", mock.Anything).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)

	pending.AssertNotCalled(t, "Release")
	peers.AssertNotCalled(t, "Add")
	rep.AssertNotCalled(t, "Success")
	conn.AssertNotCalled(t, "Close")
	rep.AssertNotCalled(t, "Failure")
	book.AssertNotCalled(t, "Block")
}

func (suite *ConnectorSuite) TestConnectorDialFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	ack := append(suite.cfg.network, nonce...)

	conn := &ConnMock{}
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(0, nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", mock.Anything)
	rep.On("Failure", mock.Anything)

	peers := &PeerManagerMock{}
	peers.On("Known", mock.Anything).Return(false)
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	dialer := &DialManagerMock{}
	dialer.On("Dial", mock.Anything).Return(nil, errors.New("could not dial address"))

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	rep.AssertCalled(t, "Failure", address)

	peers.AssertNotCalled(t, "Add", conn, nonce)
	rep.AssertNotCalled(t, "Success", address)
	conn.AssertNotCalled(t, "Close")
	book.AssertNotCalled(t, "Block")
}

func (suite *ConnectorSuite) TestConnectorWriteFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	ack := append(suite.cfg.network, nonce...)

	conn := &ConnMock{}
	conn.On("Write", mock.Anything).Return(0, errors.New("could not write syn"))
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(0, nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", mock.Anything)
	rep.On("Failure", mock.Anything)

	peers := &PeerManagerMock{}
	peers.On("Known", mock.Anything).Return(false)
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	dialer := &DialManagerMock{}
	dialer.On("Dial", mock.Anything).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	conn.AssertCalled(t, "Close")
	rep.AssertCalled(t, "Failure", address)

	peers.AssertCalled(t, "Add")
	rep.AssertCalled(t, "Success")
	book.AssertNotCalled(t, "Block")
}

func (suite *ConnectorSuite) TestConnectorReadFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	ack := append(suite.cfg.network, nonce...)

	conn := &ConnMock{}
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(0, errors.New("could not read ack"))
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", mock.Anything)
	rep.On("Failure", mock.Anything)

	peers := &PeerManagerMock{}
	peers.On("Known", mock.Anything).Return(false)
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	dialer := &DialManagerMock{}
	dialer.On("Dial", mock.Anything).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	conn.AssertCalled(t, "Close")
	rep.AssertCalled(t, "Failure", address)

	peers.AssertNotCalled(t, "Add")
	rep.AssertNotCalled(t, "Success")
	book.AssertNotCalled(t, "Block")
}

func (suite *ConnectorSuite) TestConnectorNetworkMismatch() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	ack := append([]byte{1, 2, 3, 4}, nonce...)

	conn := &ConnMock{}
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(0, nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", mock.Anything)
	rep.On("Failure", mock.Anything)

	peers := &PeerManagerMock{}
	peers.On("Known", mock.Anything).Return(false)
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	dialer := &DialManagerMock{}
	dialer.On("Dial", mock.Anything).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	conn.AssertCalled(t, "Close")
	book.AssertCalled(t, "Block", address)

	peers.AssertNotCalled(t, "Add")
	rep.AssertNotCalled(t, "Success")
	rep.AssertNotCalled(t, "Failure")
}

func (suite *ConnectorSuite) TestConnectorNonceIdentical() {

	// arrange
	address := "192.0.2.100:1337"
	ack := append(suite.cfg.network, suite.cfg.nonce...)

	conn := &ConnMock{}
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(0, nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", mock.Anything)
	rep.On("Failure", mock.Anything)

	peers := &PeerManagerMock{}
	peers.On("Known", mock.Anything).Return(false)
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	dialer := &DialManagerMock{}
	dialer.On("Dial", mock.Anything).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	conn.AssertCalled(t, "Close")
	book.AssertCalled(t, "Block", address)

	peers.AssertNotCalled(t, "Add")
	rep.AssertNotCalled(t, "Success")
	rep.AssertNotCalled(t, "Failure")
}

func (suite *ConnectorSuite) TestConnectorNonceKnown() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	ack := append(suite.cfg.network, nonce...)

	conn := &ConnMock{}
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(0, nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", mock.Anything)
	rep.On("Failure", mock.Anything)

	peers := &PeerManagerMock{}
	peers.On("Known", mock.Anything).Return(false)
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	dialer := &DialManagerMock{}
	dialer.On("Dial", mock.Anything).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	conn.AssertCalled(t, "Close")
	book.AssertCalled(t, "Block", address)

	peers.AssertNotCalled(t, "Add")
	rep.AssertNotCalled(t, "Success")
	rep.AssertNotCalled(t, "Failure")
}

func (suite *ConnectorSuite) TestConnectorAddPeerFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	ack := append(suite.cfg.network, nonce...)

	conn := &ConnMock{}
	conn.On("Write", mock.Anything).Return(0, nil)
	conn.On("Read", mock.Anything).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(0, nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", mock.Anything)
	rep.On("Failure", mock.Anything)

	peers := &PeerManagerMock{}
	peers.On("Known", mock.Anything).Return(false)
	peers.On("Add", mock.Anything, mock.Anything).Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", mock.Anything).Return(nil)
	pending.On("Release", mock.Anything).Return(nil)

	book := &AddressManagerMock{}
	book.On("Block", mock.Anything)

	dialer := &DialManagerMock{}
	dialer.On("Dial", mock.Anything).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	t := suite.T()

	pending.AssertCalled(t, "Claim", address)
	pending.AssertCalled(t, "Release", address)
	peers.AssertCalled(t, "Add", conn, nonce)
	conn.AssertCalled(t, "Close")

	book.AssertNotCalled(t, "Block")
	rep.AssertNotCalled(t, "Success")
	rep.AssertNotCalled(t, "Failure")
}

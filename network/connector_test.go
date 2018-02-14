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

func (suite *ConnectorSuite) TestConnectorClaimFails() {

	// arrange
	address := "192.0.2.100:1337"

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(errors.New("cannot claim slot"))

	peers := &PeerManagerMock{}

	rep := &ReputationManagerMock{}

	book := &AddressManagerMock{}

	dialer := &DialManagerMock{}

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertNotCalled(suite.T(), "Release", address)
}

func (suite *ConnectorSuite) TestConnectorDialFails() {

	// arrange
	address := "192.0.2.100:1337"

	rep := &ReputationManagerMock{}
	rep.On("Failure", address)

	peers := &PeerManagerMock{}

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	book := &AddressManagerMock{}

	dialer := &DialManagerMock{}
	dialer.On("Dial", address).Return(nil, errors.New("cannot dial address"))

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	rep.AssertCalled(suite.T(), "Failure", address)
}

func (suite *ConnectorSuite) TestConnectorWriteFails() {

	// arrange
	address := "192.0.2.100:1337"
	syn := append(suite.cfg.network, suite.cfg.nonce...)

	conn := &ConnMock{}
	conn.On("Close").Return(nil)
	conn.On("Write", syn).Return(0, errors.New("cannot write to connection"))

	rep := &ReputationManagerMock{}
	rep.On("Error", address)

	peers := &PeerManagerMock{}

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	book := &AddressManagerMock{}

	dialer := &DialManagerMock{}
	dialer.On("Dial", address).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	rep.AssertCalled(suite.T(), "Error", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorSuite) TestConnectorReadFails() {

	// arrange
	address := "192.0.2.100:1337"
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))

	conn := &ConnMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Return(0, errors.New("cannot read from connection"))
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Error", address)

	peers := &PeerManagerMock{}

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	book := &AddressManagerMock{}

	dialer := &DialManagerMock{}
	dialer.On("Dial", address).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	rep.AssertCalled(suite.T(), "Error", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorSuite) TestConnectorNetworkMismatch() {

	// arrange
	address := "192.0.2.100:1337"
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))
	ack := append([]byte{1, 2, 3, 4}, uuid.NewV4().Bytes()...)

	conn := &ConnMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Invalid", address)

	peers := &PeerManagerMock{}

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	book := &AddressManagerMock{}

	dialer := &DialManagerMock{}
	dialer.On("Dial", address).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	rep.AssertCalled(suite.T(), "Invalid", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorSuite) TestConnectorNonceIdentical() {

	// arrange
	address := "192.0.2.100:1337"
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, suite.cfg.nonce...)

	conn := &ConnMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Invalid", address)

	peers := &PeerManagerMock{}

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	book := &AddressManagerMock{}

	dialer := &DialManagerMock{}
	dialer.On("Dial", address).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	rep.AssertCalled(suite.T(), "Invalid", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorSuite) TestConnectorNonceKnown() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, nonce...)

	conn := &ConnMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Invalid", address)

	peers := &PeerManagerMock{}
	peers.On("Known", nonce).Return(true)

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	book := &AddressManagerMock{}

	dialer := &DialManagerMock{}
	dialer.On("Dial", address).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	rep.AssertCalled(suite.T(), "Invalid", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorSuite) TestConnectorAddPeerFails() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, nonce...)

	conn := &ConnMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Invalid", address)

	peers := &PeerManagerMock{}
	peers.On("Known", nonce).Return(false)
	peers.On("Add", conn, nonce).Return(errors.New("cannot add peer"))

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	book := &AddressManagerMock{}

	dialer := &DialManagerMock{}
	dialer.On("Dial", address).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorSuite) TestConnectorSuccess() {

	// arrange
	address := "192.0.2.100:1337"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, nonce...)

	conn := &ConnMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	rep := &ReputationManagerMock{}
	rep.On("Success", address)

	peers := &PeerManagerMock{}
	peers.On("Known", nonce).Return(false)
	peers.On("Add", conn, nonce).Return(nil)

	pending := &PendingManagerMock{}
	pending.On("Claim", address).Return(nil)
	pending.On("Release", address).Return(nil)

	book := &AddressManagerMock{}

	dialer := &DialManagerMock{}
	dialer.On("Dial", address).Return(conn, nil)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, pending, peers, rep, book, dialer, address)

	// assert
	pending.AssertCalled(suite.T(), "Claim", address)
	pending.AssertCalled(suite.T(), "Release", address)
	peers.AssertCalled(suite.T(), "Add", conn, nonce)
	rep.AssertCalled(suite.T(), "Success", address)
}

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
	"net"
	"sync"
	"testing"

	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type ConnectorTestSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *ConnectorTestSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		network: Odin,
		nonce:   uuid.NewV4().Bytes(),
	}
}

func (suite *ConnectorTestSuite) TestHandleConnectingWhenCantClaimSlot() {

	// arrange
	address := "136.44.33.12:5523"

	dial := func(string) (net.Conn, error) { return nil, nil }

	infos := &connectorInfosMock{}

	actions := &connectorActionsMock{}
	actions.On("ClaimSlot").Return(errors.New("cannot claim slot"))

	events := &connectorEventsMock{}

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, infos, actions, events, dial, address)

	// assert
	actions.AssertNotCalled(suite.T(), "ReleaseSlot")
}

func (suite *ConnectorTestSuite) TestHandleConnectingWhenAddressInvalid() {

	// arrange
	address := "not_valid"

	conn := &connMock{}

	dial := func(string) (net.Conn, error) { return conn, nil }

	infos := &connectorInfosMock{}

	actions := &connectorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot")

	events := &connectorEventsMock{}
	events.On("Invalid", address)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, infos, actions, events, dial, address)

	//Assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Invalid", address)
}

func (suite *ConnectorTestSuite) TestHandleConnectingWhenCantDialAddress() {

	// arrange
	address := "136.44.33.12:5523"

	dial := func(string) (net.Conn, error) { return nil, errors.New("cannot dial address") }

	infos := &connectorInfosMock{}

	actions := &connectorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot")

	events := &connectorEventsMock{}
	events.On("Failure", address)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, infos, actions, events, dial, address)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Failure", address)
}

func (suite *ConnectorTestSuite) TestHandleConnectingWhenCantWriteSyn() {

	// arrange
	address := "136.44.33.12:5523"
	syn := append(suite.cfg.network, suite.cfg.nonce...)

	conn := &connMock{}
	conn.On("Close").Return(nil)
	conn.On("Write", syn).Return(0, errors.New("cannot write to connection"))

	dial := func(string) (net.Conn, error) { return conn, nil }

	infos := &connectorInfosMock{}

	actions := &connectorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot")

	events := &connectorEventsMock{}
	events.On("Error", address)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, infos, actions, events, dial, address)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Error", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingWhenCantReadAck() {

	// arrange
	address := "136.44.33.12:5523"
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))

	conn := &connMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Return(0, errors.New("cannot read from connection"))
	conn.On("Close").Return(nil)

	dial := func(string) (net.Conn, error) { return conn, nil }

	infos := &connectorInfosMock{}

	actions := &connectorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot")

	events := &connectorEventsMock{}
	events.On("Error", address)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, infos, actions, events, dial, address)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Error", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingWhenNetworkMismatch() {

	// arrange
	address := "136.44.33.12:5523"
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))
	ack := append([]byte{1, 2, 3, 4}, uuid.NewV4().Bytes()...)

	conn := &connMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	dial := func(string) (net.Conn, error) { return conn, nil }

	infos := &connectorInfosMock{}

	actions := &connectorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot")

	events := &connectorEventsMock{}
	events.On("Invalid", address)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, infos, actions, events, dial, address)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Invalid", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingWhenIdenticalNonce() {

	// arrange
	address := "136.44.33.12:5523"
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, suite.cfg.nonce...)

	conn := &connMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	dial := func(string) (net.Conn, error) { return conn, nil }

	infos := &connectorInfosMock{}

	actions := &connectorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot")

	events := &connectorEventsMock{}
	events.On("Invalid", address)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, infos, actions, events, dial, address)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Invalid", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingWhenNonceAlreadyKnown() {

	// arrange
	address := "136.44.33.12:5523"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, nonce...)

	conn := &connMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	dial := func(string) (net.Conn, error) { return conn, nil }

	infos := &connectorInfosMock{}
	infos.On("KnownNonce", nonce).Return(true)

	actions := &connectorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot")

	events := &connectorEventsMock{}
	events.On("Invalid", address)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, infos, actions, events, dial, address)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Invalid", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingWhenCannotAddPeer() {

	// arrange
	address := "136.44.33.12:5523"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, nonce...)

	conn := &connMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	dial := func(string) (net.Conn, error) { return conn, nil }

	infos := &connectorInfosMock{}
	infos.On("KnownNonce", nonce).Return(false)

	actions := &connectorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot")
	actions.On("AddPeer", conn, nonce).Return(errors.New("cannot add peer"))

	events := &connectorEventsMock{}
	events.On("Invalid", address)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, infos, actions, events, dial, address)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingWhenSuccess() {

	// arrange
	address := "136.44.33.12:5523"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, nonce...)

	conn := &connMock{}
	conn.On("Write", syn).Return(len(syn), nil)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), ack)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	dial := func(string) (net.Conn, error) { return conn, nil }

	infos := &connectorInfosMock{}
	infos.On("KnownNonce", nonce).Return(false)

	actions := &connectorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot")
	actions.On("AddPeer", conn, nonce).Return(nil)

	events := &connectorEventsMock{}
	events.On("Success", address)

	// act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, infos, actions, events, dial, address)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Success", address)
}

func TestConnectorTestSuite(t *testing.T) {
	suite.Run(t, new(ConnectorTestSuite))
}

type connectorEventsMock struct {
	mock.Mock
}

func (events *connectorEventsMock) Invalid(address string) {
	events.Called(address)
}

func (events *connectorEventsMock) Error(address string) {
	events.Called(address)
}

func (events *connectorEventsMock) Success(address string) {
	events.Called(address)
}

func (events *connectorEventsMock) Failure(address string) {
	events.Called(address)
}

type connectorActionsMock struct {
	mock.Mock
}

func (actions *connectorActionsMock) ClaimSlot() error {
	args := actions.Called()
	return args.Error(0)
}

func (actions *connectorActionsMock) ReleaseSlot() {
	actions.Called()
}

func (actions *connectorActionsMock) AddPeer(conn net.Conn, nonce []byte) error {
	args := actions.Called(conn, nonce)
	return args.Error(0)
}

type connectorInfosMock struct {
	mock.Mock
}

func (infos *connectorInfosMock) KnownNonce(nonce []byte) bool {
	args := infos.Called(nonce)
	return args.Bool(0)
}

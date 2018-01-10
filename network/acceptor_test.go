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
	"time"

	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type AcceptorTestSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *AcceptorTestSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		network: Odin,
		nonce:   uuid.NewV4().Bytes(),
	}
}

func (suite *AcceptorTestSuite) TestHandleAcceptingWhenCantClaimSlot() {

	// arrange
	address := "136.44.33.12:5523"

	addr := &addrMock{}
	addr.On("String").Return(address)

	conn := &connMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Close").Return(nil)

	actions := &acceptorActionsMock{}
	actions.On("ClaimSlot").Return(errors.New("cannot claim slot"))

	events := &acceptorEventsMock{}

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, actions, events, conn)

	// assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingWhenCantReadSyn() {

	// arrange
	address := "136.44.33.12:552"
	buf := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))

	addr := &addrMock{}
	addr.On("String").Return(address)

	conn := &connMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", buf).Return(1, errors.New("cannot read from connection"))
	conn.On("Close").Return(nil)

	actions := &acceptorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot").Return(nil)

	events := &acceptorEventsMock{}
	events.On("Error", address)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, actions, events, conn)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Error", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingWhenNetworkMismatch() {

	// arrange
	address := "136.44.33.12:5523"
	syn := append([]byte{1, 2, 3, 4}, uuid.NewV4().Bytes()...)
	buf := make([]byte, len(syn))

	addr := &addrMock{}
	addr.On("String").Return(address)

	conn := &connMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	actions := &acceptorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot").Return(nil)

	events := &acceptorEventsMock{}
	events.On("Invalid", address)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, actions, events, conn)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Invalid", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingWhenIdenticalNonce() {

	// arrange
	address := "136.44.33.12:5523"
	syn := append(suite.cfg.network, suite.cfg.nonce...)
	buf := make([]byte, len(syn))

	addr := &addrMock{}
	addr.On("String").Return(address)

	conn := &connMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(len(buf), nil)
	conn.On("Close").Return(nil)

	actions := &acceptorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot").Return(nil)

	events := &acceptorEventsMock{}
	events.On("Invalid", address)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, actions, events, conn)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Invalid", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingWhenCantWriteAck() {

	// arrange
	address := "136.44.33.12:5523"
	syn := append(suite.cfg.network, uuid.NewV4().Bytes()...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, suite.cfg.nonce...)

	addr := &addrMock{}
	addr.On("String").Return(address)

	conn := &connMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(len(buf), nil)
	conn.On("Write", ack).Return(0, errors.New("cannot write to connection"))
	conn.On("Close").Return(nil)

	actions := &acceptorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot").Return(nil)

	events := &acceptorEventsMock{}
	events.On("Error", address)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, actions, events, conn)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	events.AssertCalled(suite.T(), "Error", address)
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingWhenCantAddPeer() {

	// arrange
	address := "136.44.33.12:5523"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, suite.cfg.nonce...)

	addr := &addrMock{}
	addr.On("String").Return(address)

	conn := &connMock{}
	conn.On("RemoteAddr").Return(addr)

	actions := &acceptorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot").Return(nil)
	actions.On("AddPeer", conn, nonce).Return(errors.New("cannot add peer"))

	events := &acceptorEventsMock{}
	events.On("Error", address)

	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(len(buf), nil)
	conn.On("Write", ack).Return(len(ack), nil)
	conn.On("Close").Return(nil)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, actions, events, conn)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingWhenSuccess() {

	// arrange
	address := "136.44.33.12:5523"
	nonce := uuid.NewV4().Bytes()
	syn := append(suite.cfg.network, nonce...)
	buf := make([]byte, len(syn))
	ack := append(suite.cfg.network, suite.cfg.nonce...)

	addr := &addrMock{}
	addr.On("String").Return(address)

	conn := &connMock{}
	conn.On("RemoteAddr").Return(addr)
	conn.On("Read", buf).Run(func(args mock.Arguments) {
		copy(args.Get(0).([]byte), syn)
	}).Return(len(buf), nil)
	conn.On("Write", ack).Return(len(ack), nil)
	conn.On("Close").Return(nil)

	actions := &acceptorActionsMock{}
	actions.On("ClaimSlot").Return(nil)
	actions.On("ReleaseSlot").Return(nil)
	actions.On("AddPeer", conn, nonce).Return(nil)

	events := &acceptorEventsMock{}
	events.On("Success", address)

	// act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, actions, events, conn)

	// assert
	actions.AssertCalled(suite.T(), "ReleaseSlot")
	actions.AssertCalled(suite.T(), "AddPeer", conn, nonce)
	events.AssertCalled(suite.T(), "Success", address)
}

func TestAcceptorTestSuite(t *testing.T) {
	suite.Run(t, new(AcceptorTestSuite))
}

type acceptorActionsMock struct {
	mock.Mock
}

func (actions *acceptorActionsMock) ClaimSlot() error {
	args := actions.Called()
	return args.Error(0)
}

func (actions *acceptorActionsMock) ReleaseSlot() {
	actions.Called()
}

func (actions *acceptorActionsMock) AddPeer(conn net.Conn, nonce []byte) error {
	args := actions.Called(conn, nonce)
	return args.Error(0)
}

type acceptorEventsMock struct {
	mock.Mock
}

func (events *acceptorEventsMock) Invalid(address string) {
	events.Called(address)
}

func (events *acceptorEventsMock) Error(address string) {
	events.Called(address)
}

func (events *acceptorEventsMock) Success(address string) {
	events.Called(address)
}

type connMock struct {
	mock.Mock
}

func (conn *connMock) Read(b []byte) (n int, err error) {
	args := conn.Called(b)
	return args.Int(0), args.Error(1)
}

func (conn *connMock) Write(b []byte) (n int, err error) {
	args := conn.Called(b)
	return args.Int(0), args.Error(1)
}

func (conn *connMock) Close() error {
	args := conn.Called()
	return args.Error(0)
}

func (conn *connMock) LocalAddr() net.Addr {
	args := conn.Called()
	result, _ := args.Get(0).(*addrMock)
	return result
}

func (conn *connMock) RemoteAddr() net.Addr {
	args := conn.Called()
	result, _ := args.Get(0).(*addrMock)
	return result
}

func (conn *connMock) SetDeadline(t time.Time) error {
	args := conn.Called()
	return args.Error(0)
}

func (conn *connMock) SetReadDeadline(t time.Time) error {
	args := conn.Called()
	return args.Error(0)
}

func (conn *connMock) SetWriteDeadline(t time.Time) error {
	args := conn.Called()
	return args.Error(0)
}

type addrMock struct {
	mock.Mock
}

func (addr *addrMock) Network() string {
	args := addr.Called()
	return args.String(0)
}

func (addr *addrMock) String() string {
	args := addr.Called()
	return args.String(0)
}

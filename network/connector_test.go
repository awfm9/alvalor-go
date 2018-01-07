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
	"github.com/rs/zerolog"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"net"
	"os"
	"sync"
	"testing"
)

type ConnectorTestSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *ConnectorTestSuite) SetupTest() {
	suite.log = zerolog.New(os.Stderr)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		network: Odin,
		nonce:   uuid.NewV4().Bytes(),
	}
}

func (suite *ConnectorTestSuite) TestHandleConnectingDoesNotCallReleaseSlotIfCantClaim() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	addr := "136.44.33.12:5523"

	connector.On("ClaimSlot").Return(errors.New("Can't claim slot"))

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	connector.AssertNotCalled(suite.T(), "ReleaseSlot")
}

func (suite *ConnectorTestSuite) TestHandleConnectingNotifiesBookIfAddressInvalid() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	addr := "136.44.33.1024dd"

	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	book.On("Invalid", addr)

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	book.AssertCalled(suite.T(), "Invalid", addr)
}

func (suite *ConnectorTestSuite) TestHandleConnectingNotifiesBookIfCannotDialAddress() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	book.On("Failure", addr)
	dialer.On("Dial", tcpAddr).Return(&net.TCPConn{}, errors.New("Cannot dial this address"))

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	book.AssertCalled(suite.T(), "Failure", addr)
}

func TestConnectorTestSuite(t *testing.T) {
	suite.Run(t, new(ConnectorTestSuite))
}

type connectorMock struct {
	mock.Mock
}

func (connector *connectorMock) ClaimSlot() error {
	args := connector.Called()
	return args.Error(0)
}
func (connector *connectorMock) ReleaseSlot() {
	connector.Called()
}
func (connector *connectorMock) KnownNonce(nonce []byte) bool {
	args := connector.Called(nonce)
	return args.Bool(0)
}
func (connector *connectorMock) AddPeer(conn net.Conn, nonce []byte) error {
	args := connector.Called(conn, nonce)
	return args.Error(0)
}

type tcpDialerMock struct {
	mock.Mock
}

func (dialer *tcpDialerMock) Dial(raddr *net.TCPAddr) (*net.TCPConn, error) {
	args := dialer.Called(raddr)
	return args.Get(0).(*net.TCPConn), args.Error(1)
}

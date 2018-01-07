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
	dialer.On("Dial", tcpAddr).Return(&connMock{}, errors.New("Cannot dial this address"))

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	book.AssertCalled(suite.T(), "Failure", addr)
}

func (suite *ConnectorTestSuite) TestHandleConnectingClosesConnectionIfCantWriteSyn() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	book.On("Error", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, errors.New("Can't write to this connection"))

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingNotifiesBookIfCantWriteSyn() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	book.On("Error", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, errors.New("Can't write to this connection"))

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	book.AssertCalled(suite.T(), "Error", addr)
}

func (suite *ConnectorTestSuite) TestHandleConnectingClosesConnectionIfCantReadAck() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	book.On("Error", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	conn.On("Read", make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))).Return(1, errors.New("Can't read from this connection"))

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingNotifiesBookIfCantReadAck() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	book.On("Error", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	conn.On("Read", make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))).Return(1, errors.New("Can't read from this connection"))

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	book.AssertCalled(suite.T(), "Error", addr)
}

func (suite *ConnectorTestSuite) TestHandleConnectingClosesConnectionWhenNetworkMismatch() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	book.On("Invalid", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append([]byte{66, 66, 77, 77}, uuid.NewV4().Bytes()...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingNotifiesBookWhenNetworkMismatch() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	book.On("Invalid", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append([]byte{66, 66, 77, 77}, uuid.NewV4().Bytes()...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	book.AssertCalled(suite.T(), "Invalid", addr)
}

func (suite *ConnectorTestSuite) TestHandleConnectingClosesConnectionWhenIdenticalNonce() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	book.On("Invalid", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, suite.cfg.nonce...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingNotifiesBookWhenIdenticalNonce() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)

	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	book.On("Invalid", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, suite.cfg.nonce...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	book.AssertCalled(suite.T(), "Invalid", addr)
}

func (suite *ConnectorTestSuite) TestHandleConnectingClosesConnectionWhenNonceAlreadyKnown() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	nonceIn := uuid.NewV4().Bytes()
	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	connector.On("KnownNonce", nonceIn).Return(true)
	book.On("Invalid", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, nonceIn...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *ConnectorTestSuite) TestHandleConnectingNotifiesBookWhenNonceAlreadyKnown() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	nonceIn := uuid.NewV4().Bytes()
	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	connector.On("KnownNonce", nonceIn).Return(true)
	book.On("Invalid", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, nonceIn...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	book.AssertCalled(suite.T(), "Invalid", addr)
}

func (suite *ConnectorTestSuite) TestHandleConnectingClosesConnectionWhenCannotAddPeer() {
	//Arrange
	connector := &connectorMock{}
	book := &bookMock{}
	dialer := &tcpDialerMock{}
	conn := &connMock{}
	addr := "136.44.33.15:552"
	tcpAddr, _ := net.ResolveTCPAddr("tcp", addr)
	nonceIn := uuid.NewV4().Bytes()
	connector.On("ClaimSlot").Return(nil)
	connector.On("ReleaseSlot")
	connector.On("KnownNonce", nonceIn).Return(false)
	connector.On("AddPeer", conn, nonceIn).Return(errors.New("Cannot add peer"))
	book.On("Invalid", addr)
	dialer.On("Dial", tcpAddr).Return(conn, nil)
	conn.On("Close").Return(nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, nonceIn...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)

	//Act
	handleConnecting(suite.log, &suite.wg, &suite.cfg, connector, book, dialer, addr)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
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

func (dialer *tcpDialerMock) Dial(raddr *net.TCPAddr) (net.Conn, error) {
	args := dialer.Called(raddr)
	return args.Get(0).(net.Conn), args.Error(1)
}

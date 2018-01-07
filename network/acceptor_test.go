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
	"time"
)

type AcceptorTestSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *AcceptorTestSuite) SetupTest() {
	suite.log = zerolog.New(os.Stderr)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		network:    Odin,
		listen:     false,
		address:    "0.0.0.0:31337",
		minPeers:   3,
		maxPeers:   10,
		nonce:      uuid.NewV4().Bytes(),
		interval:   time.Second * 1,
		bufferSize: 16,
	}
}

func (suite *AcceptorTestSuite) TestHandleAcceptingClosesConnectionWhenCantClaimSlot() {
	//Arrange
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(errors.New("Can't claim slot"))
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingClosesConnectionWhenCantReadSynPacket() {
	//Arrange
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	book.On("Error", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Return(1, errors.New("Can't read from connection"))
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingNotifiesBookWhenCantReadSynPacket() {
	//Arrange
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	book.On("Error", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Return(1, errors.New("Can't read from connection"))
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	book.AssertCalled(suite.T(), "Error", addr.String())
}

func (suite *AcceptorTestSuite) TestHandleAcceptingClosesConnectionWhenNetworkMismatch() {
	//Arrange
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	book.On("Invalid", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append([]byte{66, 66, 77, 77}, uuid.NewV4().Bytes()...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingNotifiesBookWhenNetworkMismatch() {
	//Arrange
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	book.On("Invalid", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append([]byte{66, 66, 77, 77}, uuid.NewV4().Bytes()...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	book.AssertCalled(suite.T(), "Invalid", addr.String())
}

func (suite *AcceptorTestSuite) TestHandleAcceptingClosesConnectionWhenIdenticalNonce() {
	//Arrange
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	book.On("Invalid", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, suite.cfg.nonce...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingNotifiesBookWhenIdenticalNonce() {
	//Arrange
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	book.On("Invalid", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, suite.cfg.nonce...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	book.AssertCalled(suite.T(), "Invalid", addr.String())
}

func (suite *AcceptorTestSuite) TestHandleAcceptingClosesConnectionWhenCantWriteAck() {
	//Arrange
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	book.On("Error", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, uuid.NewV4().Bytes()...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, errors.New("Cannot write to this tcp connection for whatever reason"))
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingNotifiesBookWhenCantWriteAck() {
	//Arrange
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	book.On("Error", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, uuid.NewV4().Bytes()...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, errors.New("Cannot write to this tcp connection for whatever reason"))
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	book.AssertCalled(suite.T(), "Error", addr.String())
}

func (suite *AcceptorTestSuite) TestHandleAcceptingClosesConnectionWhenCantAddPeer() {
	//Arrange
	nonceIn := uuid.NewV4().Bytes()
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	acceptor.On("AddPeer", conn, nonceIn).Return(errors.New("Can't add this peer"))
	book.On("Error", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, nonceIn...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	conn.AssertCalled(suite.T(), "Close")
}

func (suite *AcceptorTestSuite) TestHandleAcceptingNotifiesBookAboutSuccess() {
	//Arrange
	nonceIn := uuid.NewV4().Bytes()
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	acceptor.On("AddPeer", conn, nonceIn).Return(nil)
	book.On("Success", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, nonceIn...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	book.AssertCalled(suite.T(), "Success", addr.String())
}

func (suite *AcceptorTestSuite) TestHandleAcceptingReleasesSlotIfItClaimedBefore() {
	//Arrange
	nonceIn := uuid.NewV4().Bytes()
	acceptor := &acceptorMock{}
	book := &bookMock{}
	conn := &connMock{}
	addr := &addrMock{}
	addr.On("String").Return("136.44.33.12:5523")
	conn.On("RemoteAddr").Return(addr)

	acceptor.On("ClaimSlot").Return(nil)
	acceptor.On("ReleaseSlot").Return(nil)
	acceptor.On("AddPeer", conn, nonceIn).Return(nil)
	book.On("Success", addr.String())
	syn := make([]byte, len(suite.cfg.network)+len(suite.cfg.nonce))
	conn.On("Read", syn).Run(func(args mock.Arguments) {
		passedSyn := args.Get(0).([]byte)
		synIn := append(suite.cfg.network, nonceIn...)
		for i, val := range synIn {
			passedSyn[i] = val
		}
	}).Return(1, nil)
	conn.On("Write", append(suite.cfg.network, suite.cfg.nonce...)).Return(1, nil)
	conn.On("Close").Return(nil)

	//Act
	handleAccepting(suite.log, &suite.wg, &suite.cfg, acceptor, book, conn)

	//Assert
	acceptor.AssertCalled(suite.T(), "ReleaseSlot")
}

func TestAcceptorTestSuite(t *testing.T) {
	suite.Run(t, new(AcceptorTestSuite))
}

type acceptorMock struct {
	mock.Mock
}

func (acceptor *acceptorMock) ClaimSlot() error {
	args := acceptor.Called()
	return args.Error(0)
}
func (acceptor *acceptorMock) ReleaseSlot() {
	acceptor.Called()
}
func (acceptor *acceptorMock) AddPeer(conn net.Conn, nonce []byte) error {
	args := acceptor.Called(conn, nonce)
	return args.Error(0)
}

type bookMock struct {
	mock.Mock
}

func (book *bookMock) Add(address string) {
	book.Called(address)
}
func (book *bookMock) Invalid(address string) {
	book.Called(address)
}
func (book *bookMock) Error(address string) {
	book.Called(address)
}
func (book *bookMock) Success(address string) {
	book.Called(address)
}
func (book *bookMock) Failure(address string) {
	book.Called(address)
}
func (book *bookMock) Dropped(address string) {
	book.Called(address)
}
func (book *bookMock) Sample(count int, filter func(*Entry) bool, less func(*Entry, *Entry) bool) ([]string, error) {
	args := book.Called(count, filter, less)
	return args.Get(0).([]string), args.Error(1)
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

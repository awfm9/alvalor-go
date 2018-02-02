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
// // GNU Affero General Public License for more detailb.
// //
// // You should have received a copy of the GNU Affero General Public License
// // along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.
//
package network

//
// import (
// 	"io/ioutil"
// 	"net"
// 	"sync"
// 	"testing"
// 	"time"
//
// 	"github.com/pkg/errors"
// 	"github.com/rs/zerolog"
// 	"github.com/stretchr/testify/mock"
// 	"github.com/stretchr/testify/suite"
// )
//
// type ListenerTestSuite struct {
// 	suite.Suite
// 	log zerolog.Logger
// 	wg  sync.WaitGroup
// 	cfg Config
// }
//
// func (suite *ListenerTestSuite) SetupTest() {
// 	suite.log = zerolog.New(ioutil.Discard)
// 	suite.wg = sync.WaitGroup{}
// 	suite.wg.Add(1)
// 	suite.cfg = Config{
// 		address: "66.37.13.55:5643",
// 	}
// }
//
// func (suite *ListenerTestSuite) TestHandleListeningDoesNotStartAcceptorIfCantAcceptConnection() {
// 	// arrange
// 	conn := &connMock{}
//
// 	actions := &listenerActionsMock{}
// 	actions.On("StartAcceptor", conn)
//
// 	stop := make(chan struct{})
//
// 	listener := &listenerMock{}
// 	listener.On("SetDeadline", mock.Anything).Return(nil)
// 	//err := &net.OpError{Op: "read", Err: errors.New("Error while accepting connection")}
// 	listener.On("Accept").Return(conn, errors.New("Error while accepting connection"))
// 	listener.On("Close").Return(nil)
//
// 	go func() {
// 		time.Sleep(50 * time.Millisecond)
// 		stop <- struct{}{}
// 	}()
//
// 	// act
// 	handleListening(suite.log, &suite.wg, &suite.cfg, actions, func(string) (Listener, error) { return listener, nil }, stop)
//
// 	// assert
// 	actions.AssertNotCalled(suite.T(), "StartAcceptor", conn)
// }
//
// func (suite *ListenerTestSuite) TestHandleListeningStartsAcceptor() {
// 	// arrange
// 	conn := &connMock{}
//
// 	actions := &listenerActionsMock{}
// 	actions.On("StartAcceptor", conn)
//
// 	stop := make(chan struct{})
//
// 	listener := &listenerMock{}
// 	listener.On("SetDeadline", mock.Anything).Return(nil)
// 	listener.On("Accept").Return(conn, nil)
// 	listener.On("Close").Return(nil)
//
// 	go func() {
// 		time.Sleep(50 * time.Millisecond)
// 		stop <- struct{}{}
// 	}()
//
// 	// act
// 	handleListening(suite.log, &suite.wg, &suite.cfg, actions, func(string) (Listener, error) { return listener, nil }, stop)
//
// 	// assert
// 	actions.AssertCalled(suite.T(), "StartAcceptor", conn)
// 	listener.AssertCalled(suite.T(), "Close")
// }
//
// func TestListenerTestSuite(t *testing.T) {
// 	suite.Run(t, new(ListenerTestSuite))
// }
//
// type listenerActionsMock struct {
// 	mock.Mock
// }
//
// func (actions *listenerActionsMock) StartAcceptor(conn net.Conn) {
// 	actions.Called(conn)
// }
//
// type listenerMock struct {
// 	mock.Mock
// }
//
// func (listener *listenerMock) Accept() (net.Conn, error) {
// 	args := listener.Called()
// 	return args.Get(0).(net.Conn), args.Error(1)
// }
//
// func (listener *listenerMock) Close() error {
// 	args := listener.Called()
// 	return args.Error(0)
// }
//
// func (listener *listenerMock) SetDeadline(time time.Time) error {
// 	args := listener.Called(time)
// 	return args.Error(0)
// }

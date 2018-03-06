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

import (
	"errors"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/awishformore/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestListener(t *testing.T) {
	suite.Run(t, new(ListenerSuite))
}

type ListenerSuite struct {
	suite.Suite
	log zerolog.Logger
	wg  sync.WaitGroup
	cfg Config
}

func (suite *ListenerSuite) SetupTest() {
	suite.log = zerolog.New(ioutil.Discard)
	suite.wg = sync.WaitGroup{}
	suite.wg.Add(1)
	suite.cfg = Config{
		address: "192.0.2.100:1337",
	}
}

func (suite *ListenerSuite) TestListenerSuccess() {

	// arrange
	conn := &ConnMock{}

	ln := &ListenerMock{}
	ln.On("SetDeadline", mock.Anything).Return(nil)
	ln.On("Accept").Return(conn, nil).Once()
	ln.On("Accept").Return(nil, errors.New("could not accept connection"))
	ln.On("Close").Return(nil)

	listener := &ListenManagerMock{}
	listener.On("Listen", suite.cfg.address).Return(ln, nil)

	handlers := &HandlerManagerMock{}
	handlers.On("Acceptor", conn)

	// act
	go handleListening(suite.log, &suite.wg, &suite.cfg, handlers, listener, nil)
	suite.wg.Wait()

	// assert
	t := suite.T()

	handlers.AssertCalled(t, "Acceptor", conn)
}

func (suite *ListenerSuite) TestListenerListeningFails() {

	// arrange
	ln := &ListenerMock{}

	listener := &ListenManagerMock{}
	listener.On("Listen", suite.cfg.address).Return(ln, errors.New("could not listen"))

	handlers := &HandlerManagerMock{}
	handlers.On("Acceptor", mock.Anything)

	// act
	go handleListening(suite.log, &suite.wg, &suite.cfg, handlers, listener, nil)
	suite.wg.Wait()

	// assert
	t := suite.T()

	listener.AssertCalled(t, "Listen", suite.cfg.address)

	ln.AssertNotCalled(t, "Accept")
	ln.AssertNotCalled(t, "Close")
	handlers.AssertNotCalled(t, "Acceptor", mock.Anything)
}

func (suite *ListenerSuite) TestListenerAcceptFails() {

	// arrange
	ln := &ListenerMock{}
	ln.On("SetDeadline", mock.Anything).Return(nil)
	ln.On("Accept").Return(nil, errors.New("could not accept connection"))
	ln.On("Close").Return(nil)

	listener := &ListenManagerMock{}
	listener.On("Listen", suite.cfg.address).Return(ln, nil)

	handlers := &HandlerManagerMock{}

	// act
	go handleListening(suite.log, &suite.wg, &suite.cfg, handlers, listener, nil)
	suite.wg.Wait()

	// assert
	t := suite.T()

	handlers.AssertNotCalled(t, "Acceptor", mock.Anything)
}

func (suite *ListenerSuite) TestListenerShutdown() {

	// arrange
	stop := make(chan struct{})
	close(stop)

	ln := &ListenerMock{}
	ln.On("Close").Return(errors.New("could not close listener"))

	listener := &ListenManagerMock{}
	listener.On("Listen", suite.cfg.address).Return(ln, nil)

	handlers := &HandlerManagerMock{}

	// act
	go handleListening(suite.log, &suite.wg, &suite.cfg, handlers, listener, stop)
	suite.wg.Wait()

	// assert
	t := suite.T()

	ln.AssertCalled(t, "Close")
}

func (suite *ListenerSuite) TestListenerTimeout() {

	// arrange
	conn := &ConnMock{}

	err := &ErrorMock{}
	err.On("Error").Return("error")
	err.On("Timeout").Return(true)

	ln := &ListenerMock{}
	ln.On("SetDeadline", mock.Anything).Return(nil)
	ln.On("Accept").Return(nil, err).Once()
	ln.On("Accept").Return(conn, nil).Once()
	ln.On("Accept").Return(nil, errors.New("could not accept connection"))
	ln.On("Close").Return(nil)

	listener := &ListenManagerMock{}
	listener.On("Listen", suite.cfg.address).Return(ln, nil)

	handlers := &HandlerManagerMock{}
	handlers.On("Acceptor", conn)

	// act
	go handleListening(suite.log, &suite.wg, &suite.cfg, handlers, listener, nil)
	suite.wg.Wait()

	// assert
	t := suite.T()

	ln.AssertCalled(t, "Accept")
	handlers.AssertCalled(t, "Acceptor", conn)
}

func (suite *ListenerSuite) TestListenerCloseFails() {

	// arrange
	conn := &ConnMock{}

	ln := &ListenerMock{}
	ln.On("SetDeadline", mock.Anything).Return(nil)
	ln.On("Accept").Return(conn, nil).Once()
	ln.On("Accept").Return(nil, errors.New("could not accept connection"))
	ln.On("Close").Return(errors.New("could not close listener"))

	listener := &ListenManagerMock{}
	listener.On("Listen", suite.cfg.address).Return(ln, nil)

	handlers := &HandlerManagerMock{}
	handlers.On("Acceptor", conn)

	// act
	go handleListening(suite.log, &suite.wg, &suite.cfg, handlers, listener, nil)
	suite.wg.Wait()

	// assert
	t := suite.T()

	ln.AssertCalled(t, "Close")
}

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
	"github.com/stretchr/testify/suite"
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

func TestAcceptorTestSuite(t *testing.T) {
	suite.Run(t, new(AcceptorTestSuite))
}

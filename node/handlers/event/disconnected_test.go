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

package event

import (
	"io/ioutil"
	"sync"
	"testing"

	"github.com/alvalor/alvalor-go/network"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

func TestProcessDisconnectedSuccess(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"

	// initialize entities
	wg := &sync.WaitGroup{}
	event := network.Disconnected{Address: address}

	// initialize mocks
	net := &NetworkMock{}
	headers := &HeadersMock{}
	peers := &PeersMock{}
	message := &MessageMock{}

	// initialize handler
	handler := &Handler{
		log:     zerolog.New(ioutil.Discard),
		net:     net,
		headers: headers,
		peers:   peers,
		message: message,
	}

	// program mocks
	peers.On("Inactive", mock.Anything)

	// execute process
	handler.Process(wg, event)
	wg.Wait()

	// assert conditions
	if peers.AssertNumberOfCalls(t, "Inactive", 1) {
		peers.AssertCalled(t, "Inactive", address)
	}
}

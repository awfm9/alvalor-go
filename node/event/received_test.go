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
	"github.com/alvalor/alvalor-go/node/message"
	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

func TestProcessReceivedSuccess(t *testing.T) {

	// initialize parameters
	hash := types.Hash{0x1}
	address := "192.0.2.1"

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &message.GetTx{Hash: hash}
	event := network.Received{Address: address, Message: msg}

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
	message.On("Process", mock.Anything, mock.Anything, mock.Anything)

	// execute process
	handler.Process(wg, event)
	wg.Wait()

	// assert conditions
	if message.AssertNumberOfCalls(t, "Process", 1) {
		message.AssertCalled(t, "Process", wg, address, msg)
	}
}

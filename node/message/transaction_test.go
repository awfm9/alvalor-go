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

package message

import (
	"sync"
	"testing"

	"github.com/alvalor/alvalor-go/types"
	"github.com/stretchr/testify/mock"
)

func TestProcessTransactionSuccess(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &types.Transaction{Nonce: 1}
	msg.GetHash()

	// initialize mocks
	downloads := &DownloadsMock{}
	peers := &PeersMock{}
	entity := &EntityMock{}

	// initialize handler
	handler := &Handler{
		downloads: downloads,
		peers:     peers,
		entity:    entity,
	}

	// program mocks
	downloads.On("Cancel", mock.Anything)
	peers.On("Received", mock.Anything, mock.Anything)
	entity.On("Process", mock.Anything, mock.Anything)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	if downloads.AssertNumberOfCalls(t, "Cancel", 1) {
		downloads.AssertCalled(t, "Cancel", msg.Hash)
	}

	if peers.AssertNumberOfCalls(t, "Received", 1) {
	}

	if entity.AssertNumberOfCalls(t, "Process", 1) {
	}
}

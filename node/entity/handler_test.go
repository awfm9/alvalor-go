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

package entity

import (
	"io/ioutil"
	"sync"
	"testing"

	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

func TestHeaderSuccess(t *testing.T) {

	// initialize entities
	entity := &types.Header{Parent: types.Hash{0x1}}
	hash := entity.GetHash()
	addresses := []string{"192.0.2.1", "192.0.2.2", "192.0.2.3"}
	path := []types.Hash{{0x2}, {0x3}, {0x4}}

	// initialize mocks
	headers := &HeadersMock{}
	events := &EventsMock{}
	peers := &PeersMock{}
	net := &NetworkMock{}
	paths := &PathsMock{}

	// program mocks
	headers.On("Has", mock.Anything).Return(false)
	headers.On("Add", mock.Anything).Return(nil)
	events.On("Header", mock.Anything)
	peers.On("Addresses", mock.Anything).Return(addresses)
	net.On("Broadcast", mock.Anything, mock.Anything).Return(nil)
	headers.On("Path").Return(path, 0)
	paths.On("Follow", mock.Anything).Return(nil)

	// initialize handler
	handler := &Handler{
		log:     zerolog.New(ioutil.Discard),
		wg:      &sync.WaitGroup{},
		headers: headers,
		events:  events,
		peers:   peers,
		net:     net,
		paths:   paths,
	}

	// process entity
	handler.Process(entity)

	// assert conditions
	headers.AssertCalled(t, "Has", hash)
	headers.AssertCalled(t, "Add", entity)
	events.AssertCalled(t, "Header", hash)
	peers.AssertCalled(t, "Addresses", mock.Anything)
	net.AssertCalled(t, "Broadcast", entity, addresses)
	headers.AssertCalled(t, "Path")
	paths.AssertCalled(t, "Follow", path)
}

func TestEntityTransaction(t *testing.T) {

	// arrange
	// handler := &Handler{}
	//
	// entity := &types.Transaction{}

	// address1 := "192.0.2.1:1337"
	// address2 := "192.0.2.2:1337"
	// address3 := "192.0.2.3:1337"
	//
	// net := &NetworkMock{}
	// net.On("Send", address1, mock.Anything).Return(errors.New("could not send"))
	// net.On("Send", address2, mock.Anything).Return(nil)
	// net.On("Send", address3, mock.Anything).Return(nil)
	//
	// finder := &PathfinderMock{}
	//
	// peers := &PeersMock{}
	// peers.On("Tags", mock.Anything).Return([]string{address2})
	// peers.On("Actives").Return([]string{address1, address2, address3})
	//
	// pool := &PoolMock{}
	// pool.On("Known", mock.Anything).Return(false)
	// pool.On("Add", mock.Anything).Return(nil)
	//
	// events := &EventManagerMock{}
	// events.On("Transaction", entity.Hash).Return(nil)

	// net.AssertCalled(t, "Send", address1, mock.Anything)
	// net.AssertCalled(t, "Send", address3, mock.Anything)
	//
	// net.AssertNotCalled(t, "Send", address2, mock.Anything)
}

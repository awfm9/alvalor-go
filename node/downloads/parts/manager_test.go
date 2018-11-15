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

package parts

import (
	"errors"
	"testing"

	"github.com/alvalor/alvalor-go/node/handlers/message"
	"github.com/alvalor/alvalor-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewManager(t *testing.T) {

	// initialize mocks
	net := &NetworkMock{}
	peers := &PeersMock{}

	// initialize manager
	mgr := NewManager(net, peers)

	// check conditions
	assert.Equal(t, net, mgr.net)
	assert.Equal(t, peers, mgr.peers)
	assert.NotNil(t, mgr.pending)
}

func TestManagerStartValid(t *testing.T) {

	// initialize parameters
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"

	// initialize entities
	addresses := []string{address1, address2}
	request := &message.GetInv{Hash: hash1}

	// initialize mocks
	net := &NetworkMock{}
	peers := &PeersMock{}

	// initialize manager
	mgr := Manager{
		net:     net,
		peers:   peers,
		pending: make(map[types.Hash]string),
	}

	// initialize state
	mgr.pending[hash2] = address2
	mgr.pending[hash3] = address3

	// program mocks
	peers.On("Addresses", mock.Anything, mock.Anything).Return(addresses)
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute start
	err := mgr.Start(hash1)

	// assert conditions
	assert.Nil(t, err)

	if net.AssertNumberOfCalls(t, "Send", 1) {
		net.AssertCalled(t, "Send", address1, request)
	}

	if assert.Contains(t, mgr.pending, hash1) {
		assert.Equal(t, address1, mgr.pending[hash1])
	}
}

func TestManagerStartExisting(t *testing.T) {

	// initialize parameters
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"

	// initialize entities
	addresses := []string{address1, address2, address3}

	// initialize mocks
	net := &NetworkMock{}
	peers := &PeersMock{}

	// initialize manager
	mgr := Manager{
		net:     net,
		peers:   peers,
		pending: make(map[types.Hash]string),
	}

	// initialize state
	mgr.pending[hash1] = address1
	mgr.pending[hash2] = address2
	mgr.pending[hash3] = address3

	// program mocks
	peers.On("Addresses", mock.Anything, mock.Anything).Return(addresses)
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute start
	err := mgr.Start(hash1)

	// assert conditions
	assert.NotNil(t, err)

	net.AssertNumberOfCalls(t, "Send", 0)
}

func TestManagerStartNoPeers(t *testing.T) {

	// initialize parameters
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"

	// initialize entities

	// initialize mocks
	net := &NetworkMock{}
	peers := &PeersMock{}

	// initialize manager
	mgr := Manager{
		net:     net,
		peers:   peers,
		pending: make(map[types.Hash]string),
	}

	// initialize state
	mgr.pending[hash2] = address2
	mgr.pending[hash3] = address3

	// program mocks
	peers.On("Addresses", mock.Anything, mock.Anything).Return(nil)
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute start
	err := mgr.Start(hash1)

	// assert conditions
	assert.NotNil(t, err)

	net.AssertNumberOfCalls(t, "Send", 0)

	assert.NotContains(t, mgr.pending, hash1)
}

func TestManagerStartSendFails(t *testing.T) {

	// initialize parameters
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"

	// initialize entities
	addresses := []string{address1, address2, address3}
	request := &message.GetInv{Hash: hash1}

	// initialize mocks
	net := &NetworkMock{}
	peers := &PeersMock{}

	// initialize manager
	mgr := Manager{
		net:     net,
		peers:   peers,
		pending: make(map[types.Hash]string),
	}

	// initialize state
	mgr.pending[hash2] = address2
	mgr.pending[hash3] = address3

	// program mocks
	peers.On("Addresses", mock.Anything, mock.Anything).Return(addresses)
	net.On("Send", mock.Anything, mock.Anything).Return(errors.New(""))

	// execute start
	err := mgr.Start(hash1)

	// assert conditions
	assert.NotNil(t, err)

	if net.AssertNumberOfCalls(t, "Send", 1) {
		net.AssertCalled(t, "Send", address1, request)
	}

	assert.NotContains(t, mgr.pending, hash1)
}

func TestDownloadCancelValid(t *testing.T) {

	// initialize parameters
	hash := types.Hash{0x1}
	address := "192.0.2.1"

	// initialize mocks
	net := &NetworkMock{}
	peers := &PeersMock{}

	// initialize manager
	mgr := Manager{
		net:     net,
		peers:   peers,
		pending: make(map[types.Hash]string),
	}

	// initialize state
	mgr.pending[hash] = address

	// execute cancel
	err := mgr.Cancel(hash)

	// check conditions
	assert.Nil(t, err)

	assert.NotContains(t, mgr.pending, hash)
}

func TestDownloadCancelMissing(t *testing.T) {

	// initialize parameters
	hash := types.Hash{0x1}

	// initialize mocks
	net := &NetworkMock{}
	peers := &PeersMock{}

	// initialize manager
	mgr := Manager{
		net:     net,
		peers:   peers,
		pending: make(map[types.Hash]string),
	}

	// execute cancel
	err := mgr.Cancel(hash)

	// check conditions
	assert.NotNil(t, err)
}

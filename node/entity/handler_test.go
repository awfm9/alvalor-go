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
	"errors"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

func TestHeaderKnown(t *testing.T) {

	// initialize parameters
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}

	// initialize entities
	wg := &sync.WaitGroup{}
	entity := &types.Header{Nonce: 1}
	hash := entity.GetHash()
	addresses := []string{address1, address2, address3}
	path := []types.Hash{hash1, hash2, hash3}

	// initialize mocks
	headers := &HeadersMock{}
	events := &EventsMock{}
	peers := &PeersMock{}
	net := &NetworkMock{}
	paths := &PathsMock{}

	// program mocks
	headers.On("Has", mock.Anything).Return(true)
	headers.On("Add", mock.Anything).Return(nil)
	events.On("Header", mock.Anything)
	peers.On("Addresses", mock.Anything).Return(addresses)
	net.On("Broadcast", mock.Anything, mock.Anything).Return(nil)
	headers.On("Path").Return(path, 0)
	paths.On("Follow", mock.Anything).Return(nil)

	// initialize handler
	handler := &Handler{
		log:     zerolog.New(ioutil.Discard),
		headers: headers,
		events:  events,
		peers:   peers,
		net:     net,
		paths:   paths,
	}

	// execute process
	handler.Process(wg, entity)
	wg.Wait()

	// check conditions
	headers.AssertCalled(t, "Has", hash)
	headers.AssertNotCalled(t, "Add", mock.Anything)
	events.AssertNotCalled(t, "Header", mock.Anything)
	peers.AssertNotCalled(t, "Addresses", mock.Anything)
	net.AssertNotCalled(t, "Broadcast", mock.Anything, mock.Anything)
	headers.AssertNotCalled(t, "Path")
	paths.AssertNotCalled(t, "Follow", mock.Anything)
}

func TestHeaderAddFails(t *testing.T) {

	// initialize parameters
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}

	// initialize entities
	wg := &sync.WaitGroup{}
	entity := &types.Header{Nonce: 1}
	hash := entity.GetHash()
	addresses := []string{address1, address2, address3}
	path := []types.Hash{hash1, hash2, hash3}

	// initialize mocks
	headers := &HeadersMock{}
	events := &EventsMock{}
	peers := &PeersMock{}
	net := &NetworkMock{}
	paths := &PathsMock{}

	// program mocks
	headers.On("Has", mock.Anything).Return(false)
	headers.On("Add", mock.Anything).Return(errors.New(""))
	events.On("Header", mock.Anything)
	peers.On("Addresses", mock.Anything).Return(addresses)
	net.On("Broadcast", mock.Anything, mock.Anything).Return(nil)
	headers.On("Path").Return(path, 0)
	paths.On("Follow", mock.Anything).Return(nil)

	// initialize handler
	handler := &Handler{
		log:     zerolog.New(ioutil.Discard),
		headers: headers,
		events:  events,
		peers:   peers,
		net:     net,
		paths:   paths,
	}

	// execute process
	handler.Process(wg, entity)
	wg.Wait()

	// check conditions
	headers.AssertCalled(t, "Has", hash)
	headers.AssertCalled(t, "Add", entity)
	events.AssertNotCalled(t, "Header", mock.Anything)
	peers.AssertNotCalled(t, "Addresses", mock.Anything)
	net.AssertNotCalled(t, "Broadcast", mock.Anything, mock.Anything)
	headers.AssertNotCalled(t, "Path")
	paths.AssertNotCalled(t, "Follow", mock.Anything)
}

func TestHeaderBroadcastFails(t *testing.T) {

	// initialize parameters
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}

	// initialize entities
	wg := &sync.WaitGroup{}
	entity := &types.Header{Nonce: 1}
	hash := entity.GetHash()
	addresses := []string{address1, address2, address3}
	path := []types.Hash{hash1, hash2, hash3}

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
	net.On("Broadcast", mock.Anything, mock.Anything).Return(errors.New(""))
	headers.On("Path").Return(path, 0)
	paths.On("Follow", mock.Anything).Return(nil)

	// initialize handler
	handler := &Handler{
		log:     zerolog.New(ioutil.Discard),
		headers: headers,
		events:  events,
		peers:   peers,
		net:     net,
		paths:   paths,
	}

	// execute process
	handler.Process(wg, entity)
	wg.Wait()

	// check conditions
	headers.AssertCalled(t, "Has", hash)
	headers.AssertCalled(t, "Add", entity)
	events.AssertCalled(t, "Header", hash)
	peers.AssertCalled(t, "Addresses", mock.Anything)
	net.AssertCalled(t, "Broadcast", entity, addresses)
	headers.AssertNotCalled(t, "Path")
	paths.AssertNotCalled(t, "Follow", mock.Anything)
}

func TestHeaderFollowFails(t *testing.T) {

	// initialize parameters
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}

	// initialize entities
	wg := &sync.WaitGroup{}
	entity := &types.Header{Nonce: 1}
	hash := entity.GetHash()
	addresses := []string{address1, address2, address3}
	path := []types.Hash{hash1, hash2, hash3}

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
	paths.On("Follow", mock.Anything).Return(errors.New(""))

	// initialize handler
	handler := &Handler{
		log:     zerolog.New(ioutil.Discard),
		headers: headers,
		events:  events,
		peers:   peers,
		net:     net,
		paths:   paths,
	}

	// execute process
	handler.Process(wg, entity)
	wg.Wait()

	// check conditions
	headers.AssertCalled(t, "Has", hash)
	headers.AssertCalled(t, "Add", entity)
	events.AssertCalled(t, "Header", hash)
	peers.AssertCalled(t, "Addresses", mock.Anything)
	net.AssertCalled(t, "Broadcast", entity, addresses)
	headers.AssertCalled(t, "Path")
	paths.AssertCalled(t, "Follow", path)
}

func TestHeaderSuccess(t *testing.T) {

	// initialize parameters
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}

	// initialize entities
	wg := &sync.WaitGroup{}
	entity := &types.Header{Nonce: 1}
	hash := entity.GetHash()
	addresses := []string{address1, address2, address3}
	path := []types.Hash{hash1, hash2, hash3}

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
		headers: headers,
		events:  events,
		peers:   peers,
		net:     net,
		paths:   paths,
	}

	// execute process
	handler.Process(wg, entity)
	wg.Wait()

	// check conditions
	headers.AssertCalled(t, "Has", hash)
	headers.AssertCalled(t, "Add", entity)
	events.AssertCalled(t, "Header", hash)
	peers.AssertCalled(t, "Addresses", mock.Anything)
	net.AssertCalled(t, "Broadcast", entity, addresses)
	headers.AssertCalled(t, "Path")
	paths.AssertCalled(t, "Follow", path)
}

func TestTransactionKnown(t *testing.T) {

	// initialize parameters
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"

	// initialize entities
	wg := &sync.WaitGroup{}
	entity := &types.Transaction{Nonce: 1}
	hash := entity.GetHash()
	addresses := []string{address1, address2, address3}

	// initialize mocks
	transactions := &TransactionsMock{}
	events := &EventsMock{}
	peers := &PeersMock{}
	net := &NetworkMock{}

	// program mocks
	transactions.On("Has", mock.Anything).Return(true)
	transactions.On("Add", mock.Anything).Return(nil)
	events.On("Transaction", mock.Anything)
	peers.On("Addresses", mock.Anything).Return(addresses)
	net.On("Broadcast", mock.Anything, mock.Anything).Return(nil)

	// initialize handler
	handler := &Handler{
		log:          zerolog.New(ioutil.Discard),
		transactions: transactions,
		events:       events,
		peers:        peers,
		net:          net,
	}

	// execute process
	handler.Process(wg, entity)
	wg.Wait()

	// check conditions
	transactions.AssertCalled(t, "Has", hash)
	transactions.AssertNotCalled(t, "Add", mock.Anything)
	events.AssertNotCalled(t, "Transaction", mock.Anything)
	peers.AssertNotCalled(t, "Addresses", mock.Anything)
	net.AssertNotCalled(t, "Broadcast", mock.Anything, mock.Anything)
}

func TestTransactionAddFails(t *testing.T) {

	// initialize parameters
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"

	// initialize entities
	wg := &sync.WaitGroup{}
	entity := &types.Transaction{Nonce: 1}
	hash := entity.GetHash()
	addresses := []string{address1, address2, address3}

	// initialize mocks
	transactions := &TransactionsMock{}
	events := &EventsMock{}
	peers := &PeersMock{}
	net := &NetworkMock{}

	// program mocks
	transactions.On("Has", mock.Anything).Return(false)
	transactions.On("Add", mock.Anything).Return(errors.New(""))
	events.On("Transaction", mock.Anything)
	peers.On("Addresses", mock.Anything).Return(addresses)
	net.On("Broadcast", mock.Anything, mock.Anything).Return(nil)

	// initialize handler
	handler := &Handler{
		log:          zerolog.New(ioutil.Discard),
		transactions: transactions,
		events:       events,
		peers:        peers,
		net:          net,
	}

	// execute process
	handler.Process(wg, entity)
	wg.Wait()

	// check conditions
	transactions.AssertCalled(t, "Has", hash)
	transactions.AssertCalled(t, "Add", entity)
	events.AssertNotCalled(t, "Transaction", mock.Anything)
	peers.AssertNotCalled(t, "Addresses", mock.Anything)
	net.AssertNotCalled(t, "Broadcast", mock.Anything, mock.Anything)
}

func TestTransactionBroadcastFails(t *testing.T) {

	// initialize parameters
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"

	// initialize entities
	wg := &sync.WaitGroup{}
	entity := &types.Transaction{Nonce: 1}
	hash := entity.GetHash()
	addresses := []string{address1, address2, address3}

	// initialize mocks
	transactions := &TransactionsMock{}
	events := &EventsMock{}
	peers := &PeersMock{}
	net := &NetworkMock{}

	// program mocks
	transactions.On("Has", mock.Anything).Return(false)
	transactions.On("Add", mock.Anything).Return(nil)
	events.On("Transaction", mock.Anything)
	peers.On("Addresses", mock.Anything).Return(addresses)
	net.On("Broadcast", mock.Anything, mock.Anything).Return(errors.New(""))

	// initialize handler
	handler := &Handler{
		log:          zerolog.New(ioutil.Discard),
		transactions: transactions,
		events:       events,
		peers:        peers,
		net:          net,
	}

	// execute process
	handler.Process(wg, entity)
	wg.Wait()

	// check conditions
	transactions.AssertCalled(t, "Has", hash)
	transactions.AssertCalled(t, "Add", entity)
	events.AssertCalled(t, "Transaction", hash)
	peers.AssertCalled(t, "Addresses", mock.Anything)
	net.AssertCalled(t, "Broadcast", entity, addresses)
}

func TestTransactionSuccess(t *testing.T) {

	// initialize parameters
	address1 := "192.0.2.1"
	address2 := "192.0.2.2"
	address3 := "192.0.2.3"

	// initialize entities
	wg := &sync.WaitGroup{}
	entity := &types.Transaction{Nonce: 1}
	hash := entity.GetHash()
	addresses := []string{address1, address2, address3}

	// initialize mocks
	transactions := &TransactionsMock{}
	events := &EventsMock{}
	peers := &PeersMock{}
	net := &NetworkMock{}

	// program mocks
	transactions.On("Has", mock.Anything).Return(false)
	transactions.On("Add", mock.Anything).Return(nil)
	events.On("Transaction", mock.Anything)
	peers.On("Addresses", mock.Anything).Return(addresses)
	net.On("Broadcast", mock.Anything, mock.Anything).Return(nil)

	// initialize handler
	handler := &Handler{
		log:          zerolog.New(ioutil.Discard),
		transactions: transactions,
		events:       events,
		peers:        peers,
		net:          net,
	}

	// execute process
	handler.Process(wg, entity)
	wg.Wait()

	// check conditions
	transactions.AssertCalled(t, "Has", hash)
	transactions.AssertCalled(t, "Add", entity)
	events.AssertCalled(t, "Transaction", hash)
	peers.AssertCalled(t, "Addresses", mock.Anything)
	net.AssertCalled(t, "Broadcast", entity, addresses)
}

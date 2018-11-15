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
	if transactions.AssertNumberOfCalls(t, "Has", 1) {
		transactions.AssertCalled(t, "Has", hash)
	}

	transactions.AssertNumberOfCalls(t, "Add", 0)

	events.AssertNumberOfCalls(t, "Transaction", 0)

	peers.AssertNumberOfCalls(t, "Addresses", 0)

	net.AssertNumberOfCalls(t, "Broadcast", 0)
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
	if transactions.AssertNumberOfCalls(t, "Has", 1) {
		transactions.AssertCalled(t, "Has", hash)
	}

	if transactions.AssertNumberOfCalls(t, "Add", 1) {
		transactions.AssertCalled(t, "Add", entity)
	}

	events.AssertNumberOfCalls(t, "Transaction", 0)

	peers.AssertNumberOfCalls(t, "Addresses", 0)

	net.AssertNumberOfCalls(t, "Broadcast", 0)
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
	if transactions.AssertNumberOfCalls(t, "Has", 1) {
		transactions.AssertCalled(t, "Has", hash)
	}

	if transactions.AssertNumberOfCalls(t, "Add", 1) {
		transactions.AssertCalled(t, "Add", entity)
	}

	if events.AssertNumberOfCalls(t, "Transaction", 1) {
		events.AssertCalled(t, "Transaction", hash)
	}

	if peers.AssertNumberOfCalls(t, "Addresses", 1) {
		peers.AssertCalled(t, "Addresses", mock.Anything)
	}

	if net.AssertNumberOfCalls(t, "Broadcast", 1) {
		net.AssertCalled(t, "Broadcast", entity, addresses)
	}
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
	if transactions.AssertNumberOfCalls(t, "Has", 1) {
		transactions.AssertCalled(t, "Has", hash)
	}

	if transactions.AssertNumberOfCalls(t, "Add", 1) {
		transactions.AssertCalled(t, "Add", entity)
	}

	if events.AssertNumberOfCalls(t, "Transaction", 1) {
		events.AssertCalled(t, "Transaction", hash)
	}

	if peers.AssertNumberOfCalls(t, "Addresses", 1) {
		peers.AssertCalled(t, "Addresses", mock.Anything)
	}

	if net.AssertNumberOfCalls(t, "Broadcast", 1) {
		net.AssertCalled(t, "Broadcast", entity, addresses)
	}
}

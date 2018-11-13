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
	"errors"
	"io/ioutil"
	"sync"
	"testing"

	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
)

func TestProcessGetTxSuccess(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash := types.Hash{0x1}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &GetTx{Hash: hash}
	tx := &types.Transaction{Nonce: 1}

	// initialize mocks
	transactions := &TransactionsMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		log:          zerolog.New(ioutil.Discard),
		transactions: transactions,
		net:          net,
	}

	// program mocks
	transactions.On("Get", mock.Anything).Return(tx, nil)
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	transactions.AssertCalled(t, "Get", hash)
	net.AssertCalled(t, "Send", address, tx)
}

func TestProcessGetTxGetFails(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash := types.Hash{0x1}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &GetTx{Hash: hash}
	tx := &types.Transaction{Nonce: 1}

	// initialize mocks
	transactions := &TransactionsMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		log:          zerolog.New(ioutil.Discard),
		transactions: transactions,
		net:          net,
	}

	// program mocks
	transactions.On("Get", mock.Anything).Return(tx, errors.New(""))
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	transactions.AssertCalled(t, "Get", hash)
	net.AssertNotCalled(t, "Send")
}

func TestProcessGetTxSendFails(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash := types.Hash{0x1}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &GetTx{Hash: hash}
	tx := &types.Transaction{Nonce: 1}

	// initialize mocks
	transactions := &TransactionsMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		log:          zerolog.New(ioutil.Discard),
		transactions: transactions,
		net:          net,
	}

	// program mocks
	transactions.On("Get", mock.Anything).Return(tx, nil)
	net.On("Send", mock.Anything, mock.Anything).Return(errors.New(""))

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	transactions.AssertCalled(t, "Get", hash)
	net.AssertCalled(t, "Send", address, tx)
}

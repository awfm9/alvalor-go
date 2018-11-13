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

func TestProcessGetInvSuccess(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash := types.Hash{0x1}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &GetInv{Hash: hash}
	inv := &types.Inventory{Hash: hash}

	// initialize mocks
	inventories := &InventoriesMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		log:         zerolog.New(ioutil.Discard),
		inventories: inventories,
		net:         net,
	}

	// program mocks
	inventories.On("Get", mock.Anything).Return(inv, nil)
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	inventories.AssertCalled(t, "Get", hash)
	net.AssertCalled(t, "Send", address, inv)
}

func TestProcessGetInvGetFails(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash := types.Hash{0x1}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &GetInv{Hash: hash}
	inv := &types.Inventory{Hash: hash}

	// initialize mocks
	inventories := &InventoriesMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		log:         zerolog.New(ioutil.Discard),
		inventories: inventories,
		net:         net,
	}

	// program mocks
	inventories.On("Get", mock.Anything).Return(inv, errors.New(""))
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	inventories.AssertCalled(t, "Get", hash)
	net.AssertNotCalled(t, "Send")
}

func TestProcessGetInvSendFails(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash := types.Hash{0x1}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &GetInv{Hash: hash}
	inv := &types.Inventory{Hash: hash}

	// initialize mocks
	inventories := &InventoriesMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		log:         zerolog.New(ioutil.Discard),
		inventories: inventories,
		net:         net,
	}

	// program mocks
	inventories.On("Get", mock.Anything).Return(inv, nil)
	net.On("Send", mock.Anything, mock.Anything).Return(errors.New(""))

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	inventories.AssertCalled(t, "Get", hash)
	net.AssertCalled(t, "Send", address, inv)
}

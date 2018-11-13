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
	"sync"
	"testing"

	"github.com/alvalor/alvalor-go/types"
	"github.com/stretchr/testify/mock"
)

func TestProcessStatusSuccess(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	distance1 := 10
	distance2 := 20
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &Status{Distance: uint64(distance2)}
	path := []types.Hash{hash1, hash2, hash3}
	sync := &Sync{Locators: []types.Hash{hash1, hash2, hash3}}

	// initialize mocks
	headers := &HeadersMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		headers: headers,
		net:     net,
	}

	// program mocks
	headers.On("Path").Return(path, distance1)
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	headers.AssertNumberOfCalls(t, "Path", 1)

	if net.AssertNumberOfCalls(t, "Send", 1) {
		net.AssertCalled(t, "Send", address, sync)
	}
}

func TestProcessStatusBehind(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	distance1 := 10
	distance2 := 00
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &Status{Distance: uint64(distance2)}
	path := []types.Hash{hash1, hash2, hash3}

	// initialize mocks
	headers := &HeadersMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		headers: headers,
		net:     net,
	}

	// program mocks
	headers.On("Path").Return(path, distance1)
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	headers.AssertNumberOfCalls(t, "Path", 1)

	net.AssertNumberOfCalls(t, "Send", 0)
}

func TestProcessStatusSendFails(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	distance1 := 10
	distance2 := 20
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &Status{Distance: uint64(distance2)}
	path := []types.Hash{hash1, hash2, hash3}
	sync := &Sync{Locators: []types.Hash{hash1, hash2, hash3}}

	// initialize mocks
	headers := &HeadersMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		headers: headers,
		net:     net,
	}

	// program mocks
	headers.On("Path").Return(path, distance1)
	net.On("Send", mock.Anything, mock.Anything).Return(errors.New(""))

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	headers.AssertNumberOfCalls(t, "Path", 1)

	if net.AssertNumberOfCalls(t, "Send", 1) {
		net.AssertCalled(t, "Send", address, sync)
	}
}

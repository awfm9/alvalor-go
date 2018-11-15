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

func TestProcessSyncSuccess(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}
	hash4 := types.Hash{0x4}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &Sync{Locators: []types.Hash{hash1, hash2}}
	path := []types.Hash{hash4, hash3, hash2, hash1}
	header1 := &types.Header{Nonce: 1}
	header2 := &types.Header{Nonce: 2}
	header3 := &types.Header{Nonce: 3}
	header4 := &types.Header{Nonce: 4}
	pathMsg := &Path{Headers: []*types.Header{header3, header4}}

	// initialize mocks
	headers := &HeadersMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		headers: headers,
		net:     net,
	}

	// program mocks
	headers.On("Path").Return(path, 0)
	headers.On("Get", hash1).Return(header1, nil)
	headers.On("Get", hash2).Return(header2, nil)
	headers.On("Get", hash3).Return(header3, nil)
	headers.On("Get", hash4).Return(header4, nil)
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	headers.AssertNumberOfCalls(t, "Path", 1)

	if headers.AssertNumberOfCalls(t, "Get", 2) {
		headers.AssertCalled(t, "Get", hash3)
		headers.AssertCalled(t, "Get", hash4)
	}

	if net.AssertNumberOfCalls(t, "Send", 1) {
		net.AssertCalled(t, "Send", address, pathMsg)
	}
}

func TestProcessSyncNoPath(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}
	hash4 := types.Hash{0x4}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &Sync{Locators: []types.Hash{hash1, hash2}}
	path := []types.Hash{hash2, hash1}
	header1 := &types.Header{Nonce: 1}
	header2 := &types.Header{Nonce: 2}
	header3 := &types.Header{Nonce: 3}
	header4 := &types.Header{Nonce: 4}

	// initialize mocks
	headers := &HeadersMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		headers: headers,
		net:     net,
	}

	// program mocks
	headers.On("Path").Return(path, 0)
	headers.On("Get", hash1).Return(header1, nil)
	headers.On("Get", hash2).Return(header2, nil)
	headers.On("Get", hash3).Return(header3, nil)
	headers.On("Get", hash4).Return(header4, nil)
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	headers.AssertNumberOfCalls(t, "Path", 1)

	headers.AssertNumberOfCalls(t, "Get", 0)

	net.AssertNumberOfCalls(t, "Send", 0)
}

func TestProcessSyncGetFails(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}
	hash4 := types.Hash{0x4}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &Sync{Locators: []types.Hash{hash1, hash2}}
	path := []types.Hash{hash4, hash3, hash2, hash1}
	header1 := &types.Header{Nonce: 1}
	header2 := &types.Header{Nonce: 2}
	header3 := &types.Header{Nonce: 3}
	header4 := &types.Header{Nonce: 4}

	// initialize mocks
	headers := &HeadersMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		headers: headers,
		net:     net,
	}

	// program mocks
	headers.On("Path").Return(path, 0)
	headers.On("Get", hash1).Return(header1, nil)
	headers.On("Get", hash2).Return(header2, nil)
	headers.On("Get", hash3).Return(header3, errors.New(""))
	headers.On("Get", hash4).Return(header4, errors.New(""))
	net.On("Send", mock.Anything, mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	headers.AssertNumberOfCalls(t, "Path", 1)

	if headers.AssertNumberOfCalls(t, "Get", 1) {
		headers.AssertCalled(t, "Get", hash3)
	}

	net.AssertNumberOfCalls(t, "Send", 0)
}

func TestProcessSyncSendFails(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}
	hash3 := types.Hash{0x3}
	hash4 := types.Hash{0x4}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &Sync{Locators: []types.Hash{hash1, hash2}}
	path := []types.Hash{hash4, hash3, hash2, hash1}
	header1 := &types.Header{Nonce: 1}
	header2 := &types.Header{Nonce: 2}
	header3 := &types.Header{Nonce: 3}
	header4 := &types.Header{Nonce: 4}
	pathMsg := &Path{Headers: []*types.Header{header3, header4}}

	// initialize mocks
	headers := &HeadersMock{}
	net := &NetworkMock{}

	// initialize handler
	handler := &Handler{
		headers: headers,
		net:     net,
	}

	// program mocks
	headers.On("Path").Return(path, 0)
	headers.On("Get", hash1).Return(header1, nil)
	headers.On("Get", hash2).Return(header2, nil)
	headers.On("Get", hash3).Return(header3, nil)
	headers.On("Get", hash4).Return(header4, nil)
	net.On("Send", mock.Anything, mock.Anything).Return(errors.New(""))

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	headers.AssertNumberOfCalls(t, "Path", 1)

	if headers.AssertNumberOfCalls(t, "Get", 2) {
		headers.AssertCalled(t, "Get", hash3)
		headers.AssertCalled(t, "Get", hash4)
	}

	if net.AssertNumberOfCalls(t, "Send", 1) {
		net.AssertCalled(t, "Send", address, pathMsg)
	}
}

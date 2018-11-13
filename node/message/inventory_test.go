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

func TestProcessInventorySuccess(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash := types.Hash{0x1}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &types.Inventory{Hash: hash}

	// initialize mocks
	downloads := &DownloadsMock{}
	peers := &PeersMock{}
	inventories := &InventoriesMock{}
	paths := &PathsMock{}

	// initialize handler
	handler := &Handler{
		log:         zerolog.New(ioutil.Discard),
		downloads:   downloads,
		peers:       peers,
		inventories: inventories,
		paths:       paths,
	}

	// program mocks
	downloads.On("Cancel", mock.Anything)
	peers.On("Received", mock.Anything, mock.Anything)
	inventories.On("Add", mock.Anything).Return(nil)
	paths.On("Signal", mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	if downloads.AssertNumberOfCalls(t, "Cancel", 1) {
		downloads.AssertCalled(t, "Cancel", hash)
	}

	if peers.AssertNumberOfCalls(t, "Received", 1) {
		peers.AssertCalled(t, "Received", address, hash)
	}

	if inventories.AssertNumberOfCalls(t, "Add", 1) {
		inventories.AssertCalled(t, "Add", msg)
	}

	if paths.AssertNumberOfCalls(t, "Signal", 1) {
		paths.AssertCalled(t, "Signal", hash)
	}
}

func TestProcessInventoryAddFails(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash := types.Hash{0x1}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &types.Inventory{Hash: hash}

	// initialize mocks
	downloads := &DownloadsMock{}
	peers := &PeersMock{}
	inventories := &InventoriesMock{}
	paths := &PathsMock{}

	// initialize handler
	handler := &Handler{
		log:         zerolog.New(ioutil.Discard),
		downloads:   downloads,
		peers:       peers,
		inventories: inventories,
		paths:       paths,
	}

	// program mocks
	downloads.On("Cancel", mock.Anything)
	peers.On("Received", mock.Anything, mock.Anything)
	inventories.On("Add", mock.Anything).Return(errors.New(""))
	paths.On("Signal", mock.Anything).Return(nil)

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	if downloads.AssertNumberOfCalls(t, "Cancel", 1) {
		downloads.AssertCalled(t, "Cancel", hash)
	}

	if peers.AssertNumberOfCalls(t, "Received", 1) {
		peers.AssertCalled(t, "Received", address, hash)
	}

	if inventories.AssertNumberOfCalls(t, "Add", 1) {
		inventories.AssertCalled(t, "Add", msg)
	}

	paths.AssertNumberOfCalls(t, "Signal", 0)
}

func TestProcessInventorySignalFails(t *testing.T) {

	// initialize parameters
	address := "192.0.2.1"
	hash := types.Hash{0x1}

	// initialize entities
	wg := &sync.WaitGroup{}
	msg := &types.Inventory{Hash: hash}

	// initialize mocks
	downloads := &DownloadsMock{}
	peers := &PeersMock{}
	inventories := &InventoriesMock{}
	paths := &PathsMock{}

	// initialize handler
	handler := &Handler{
		log:         zerolog.New(ioutil.Discard),
		downloads:   downloads,
		peers:       peers,
		inventories: inventories,
		paths:       paths,
	}

	// program mocks
	downloads.On("Cancel", mock.Anything)
	peers.On("Received", mock.Anything, mock.Anything)
	inventories.On("Add", mock.Anything).Return(nil)
	paths.On("Signal", mock.Anything).Return(errors.New(""))

	// execute process
	handler.Process(wg, address, msg)
	wg.Wait()

	// check conditions
	if downloads.AssertNumberOfCalls(t, "Cancel", 1) {
		downloads.AssertCalled(t, "Cancel", hash)
	}

	if peers.AssertNumberOfCalls(t, "Received", 1) {
		peers.AssertCalled(t, "Received", address, hash)
	}

	if inventories.AssertNumberOfCalls(t, "Add", 1) {
		inventories.AssertCalled(t, "Add", msg)
	}

	if paths.AssertNumberOfCalls(t, "Signal", 1) {
		paths.AssertCalled(t, "Signal", hash)
	}
}

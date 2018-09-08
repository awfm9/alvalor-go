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

package node

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/alvalor/alvalor-go/types"
)

func TestBlockRequestsQueueReturnsEmptyArray(t *testing.T) {
	//Arrange
	requests := []blockRequestMessage{}
	queue := newBlockRequestsQueue(requests)

	//Act
	queueData := queue.getData()

	//Assert
	assert.Empty(t, queueData)
}

func TestBlockRequestsQueueBalanceSimpleCase(t *testing.T) {
	//Arrange
	requests := []blockRequestMessage{}
	hash1 := types.Hash{11, 12, 13}
	hash2 := types.Hash{25, 25, 35}
	hash3 := types.Hash{35, 44, 32}
	requests = append(requests, blockRequestMessage{hash: hash1, addresses: []string{"192.168.3.22", "76.33.22.71"}})
	requests = append(requests, blockRequestMessage{hash: hash2, addresses: []string{"192.168.3.22", "92.25.72.34"}})
	requests = append(requests, blockRequestMessage{hash: hash3, addresses: []string{"192.168.3.22", "23.29.113.112"}})
	queue := newBlockRequestsQueue(requests)

	//Act
	queueData := queue.getData()

	//Assert
	assert.Len(t, queueData, 3)
	assert.Equal(t, hash1, queueData["192.168.3.22"][0])
	assert.Equal(t, hash2, queueData["92.25.72.34"][0])
	assert.Equal(t, hash3, queueData["23.29.113.112"][0])
}

func TestBlockRequestsQueueBalanceWhenAllPeersHaveAllBlocks(t *testing.T) {
	//Arrange
	requests := []blockRequestMessage{}
	hash1 := types.Hash{11, 12, 13}
	hash2 := types.Hash{25, 25, 35}
	hash3 := types.Hash{35, 44, 32}
	hash4 := types.Hash{15, 15, 15}
	hash5 := types.Hash{25, 35, 55}
	hash6 := types.Hash{97, 62, 33}
	requests = append(requests, blockRequestMessage{hash: hash1, addresses: []string{"192.168.3.22", "76.33.22.71", "92.25.72.34", "23.29.113.112", "92.92.25.25", "25.25.92.92"}})
	requests = append(requests, blockRequestMessage{hash: hash2, addresses: []string{"192.168.3.22", "76.33.22.71", "92.25.72.34", "23.29.113.112", "92.92.25.25", "25.25.92.92"}})
	requests = append(requests, blockRequestMessage{hash: hash3, addresses: []string{"192.168.3.22", "76.33.22.71", "92.25.72.34", "23.29.113.112", "92.92.25.25", "25.25.92.92"}})
	requests = append(requests, blockRequestMessage{hash: hash4, addresses: []string{"192.168.3.22", "76.33.22.71", "92.25.72.34", "23.29.113.112", "92.92.25.25", "25.25.92.92"}})
	requests = append(requests, blockRequestMessage{hash: hash5, addresses: []string{"192.168.3.22", "76.33.22.71", "92.25.72.34", "23.29.113.112", "92.92.25.25", "25.25.92.92"}})
	requests = append(requests, blockRequestMessage{hash: hash6, addresses: []string{"192.168.3.22", "76.33.22.71", "92.25.72.34", "23.29.113.112", "92.92.25.25", "25.25.92.92"}})
	queue := newBlockRequestsQueue(requests)

	//Act
	queueData := queue.getData()

	//Assert
	assert.Len(t, queueData, 6)
	assert.Equal(t, hash1, queueData["192.168.3.22"][0])
	assert.Equal(t, hash2, queueData["76.33.22.71"][0])
	assert.Equal(t, hash3, queueData["92.25.72.34"][0])
	assert.Equal(t, hash4, queueData["23.29.113.112"][0])
	assert.Equal(t, hash5, queueData["92.92.25.25"][0])
	assert.Equal(t, hash6, queueData["25.25.92.92"][0])
}

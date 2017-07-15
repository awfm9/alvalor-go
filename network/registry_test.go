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

package network

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAddSavesPeer(t *testing.T) {
	reg := registry{ peers : make(map[string]*peer)}
    addr := "127.0.0.1"

	reg.add(addr, &peer{})

	assert.True(t, reg.has(addr))
}

func TestRemovesPeer(t *testing.T) {
	reg := registry{ peers : make(map[string]*peer)}
    addr := "127.0.0.1"
	reg.add(addr, &peer{})

	reg.remove(addr)

	assert.False(t, reg.has(addr))
}

func TestCountsPeers(t *testing.T) {
	reg := registry{ peers : make(map[string]*peer)}

	reg.add("127.0.0.1", &peer{})
	reg.add("192.168.4.62", &peer{})
	reg.add("192.168.77.55", &peer{})

	count := reg.count()

	assert.Equal(t, 3, count)
}

func TestGetsPeer(t *testing.T) {
	reg := registry{ peers : make(map[string]*peer)}
	addr := "127.0.0.1"
	peerToAdd := &peer{}
	reg.add(addr, peerToAdd)

	peer := reg.get(addr)

	assert.Equal(t, peerToAdd, peer)
}

func TestSlice(t *testing.T) {
	reg := registry{ peers : make(map[string]*peer)}
	peers := make([]*peer, 2)
	peers[0] = &peer{}
	peers[1] = &peer{}

	reg.add("192.168.66.22", peers[0])
	reg.add("192.168.46.84", peers[1])

	slice := reg.slice()

	assert.EqualValues(t, peers, slice)
}
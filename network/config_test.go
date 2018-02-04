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

func TestSetNetwork(t *testing.T) {
	cfg := &Config{network: []byte{0}}
	network := []byte{1}
	SetNetwork(network)(cfg)
	assert.Equal(t, network, cfg.network, "Set network did not set network")
}

func TestSetListen(t *testing.T) {
	cfg := &Config{listen: false}
	listen := true
	SetListen(listen)(cfg)
	assert.Equal(t, listen, cfg.listen, "Set listen did not set listen")
}

func TestSetAddress(t *testing.T) {
	cfg := &Config{address: "192.0.2.100:1337"}
	address := "192.0.2.200:1337"
	SetAddress(address)(cfg)
	assert.Equal(t, address, cfg.address, "Set address did not set address")
}

func TestMinPeers(t *testing.T) {
	cfg := &Config{minPeers: 0}
	minPeers := uint(1)
	SetMinPeers(minPeers)(cfg)
	assert.Equal(t, minPeers, cfg.minPeers, "Set min peers did not set min peers")
}

func TestMaxPeers(t *testing.T) {
	cfg := &Config{minPeers: 0}
	maxPeers := uint(1)
	SetMaxPeers(maxPeers)(cfg)
	assert.Equal(t, maxPeers, cfg.maxPeers, "Set max peers did not set max peers")
}

func TestMaxPending(t *testing.T) {
	cfg := &Config{maxPending: 0}
	maxPending := uint(1)
	SetMaxPending(maxPending)(cfg)
	assert.Equal(t, maxPending, cfg.maxPending, "Set max pending did not set max pending")
}

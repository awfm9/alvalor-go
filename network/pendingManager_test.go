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

func TestNewPendingManager(t *testing.T) {
	max := uint(7)
	pending := newSimplePendingManager(7)
	assert.Equal(t, max, pending.max)
	assert.NotNil(t, pending.addresses)
}

func TestPendingManagerClaim(t *testing.T) {
	address := "192.0.2.100:1337"
	pending := simplePendingManager{addresses: make(map[string]struct{})}

	pending.max = 0
	err := pending.Claim(address)
	assert.NotNil(t, err)
	assert.Len(t, pending.addresses, 0)

	pending.max = 2
	err = pending.Claim(address)
	assert.Nil(t, err)
	assert.Len(t, pending.addresses, 1)
	assert.Contains(t, pending.addresses, address)

	err = pending.Claim(address)
	assert.NotNil(t, err)
	assert.Len(t, pending.addresses, 1)
}

func TestPendingManagerRelease(t *testing.T) {
	address := "192.0.2.100:1337"
	pending := simplePendingManager{addresses: make(map[string]struct{})}

	pending.addresses[address] = struct{}{}
	err := pending.Release("other address")
	assert.NotNil(t, err)
	assert.Len(t, pending.addresses, 1)
	assert.Contains(t, pending.addresses, address)

	err = pending.Release(address)
	assert.Nil(t, err)
	assert.Len(t, pending.addresses, 0)
}

func TestPendingManagerAddresses(t *testing.T) {
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	pending := simplePendingManager{addresses: make(map[string]struct{})}

	addresses := pending.Addresses()
	assert.Empty(t, addresses)

	pending.addresses[address1] = struct{}{}
	pending.addresses[address2] = struct{}{}
	addresses = pending.Addresses()
	assert.ElementsMatch(t, []string{address1, address2}, addresses)
}

func TestPendingManagerCount(t *testing.T) {
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	pending := simplePendingManager{addresses: make(map[string]struct{})}

	count := pending.Count()
	assert.Equal(t, uint(0), count)

	pending.addresses[address1] = struct{}{}
	pending.addresses[address2] = struct{}{}
	count = pending.Count()
	assert.Equal(t, uint(2), count)
}

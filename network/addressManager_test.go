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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAddressManager(t *testing.T) {
	am := newSimpleAddressManager()
	assert.NotNil(t, am.blacklist)
	assert.NotNil(t, am.addresses)
}

func TestAddressManagerAdd(t *testing.T) {
	address := "192.0.2.100:1337"
	am := simpleAddressManager{addresses: make(map[string]bool)}
	am.Add(address)
	assert.Contains(t, am.addresses, address)
}

func TestAddressManagerRemove(t *testing.T) {
	address := "192.0.2.100:1337"
	am := simpleAddressManager{addresses: map[string]bool{address: true}}
	am.Remove(address)
	assert.NotContains(t, am.addresses, address)
}

func TestAddressManagerBlock(t *testing.T) {
	address := "192.0.2.100:1337"
	am := simpleAddressManager{blacklist: make(map[string]bool)}
	am.Block(address)
	assert.Contains(t, am.blacklist, address)
}

func TestAddressManagerUnblock(t *testing.T) {
	address := "192.0.2.100:1337"
	am := simpleAddressManager{blacklist: map[string]bool{address: true}}
	am.Unblock(address)
	assert.NotContains(t, am.blacklist, address)
}

func TestAddressManagerSample(t *testing.T) {
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.101:1337" // blacklist + filter
	address3 := "192.0.2.102:1337" // blacklist
	address4 := "192.0.2.103:1337" // filter
	address5 := "192.0.2.104:1337"
	address6 := "192.0.2.106:1337"
	address7 := "192.0.2.107:1337"
	am := simpleAddressManager{
		blacklist: map[string]bool{
			address2: true,
			address3: true,
		},
		addresses: map[string]bool{
			address1: true,
			address2: true,
			address3: true,
			address4: true,
			address5: true,
			address6: true,
			address7: true,
		},
	}
	// this will make sure we cover both less cases with high probability
	for i := 0; i < 146; i++ {
		last := 108 + i
		address := fmt.Sprintf("192.0.2.%v:1337", last)
		am.addresses[address] = true
	}
	filter := func(a string) bool {
		if a == address4 || a == address2 {
			return false
		}
		return true
	}
	less := func(a1 string, a2 string) bool {
		if strings.Compare(a1, a2) < 0 {
			return true
		}
		return false
	}
	expected := []string{
		address1,
		address5,
		address6,
		address7,
	}
	sample := am.Sample(4, filter, less)
	assert.Equal(t, expected, sample)
	sample = am.Sample(1, filter)
	assert.Contains(t, am.addresses, sample[0])
}

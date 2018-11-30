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

package peers

import (
	"testing"

	"github.com/alvalor/alvalor-go/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	state := NewState()
	assert.NotNil(t, state.peers)
}

func TestStateActive(t *testing.T) {
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	state := &State{peers: make(map[string]*Peer)}

	state.Active(address1)
	if assert.Len(t, state.peers, 1) {
		if assert.Contains(t, state.peers, address1) {
			p := state.peers[address1]
			assert.True(t, p.active)
		}
	}

	state.peers[address2] = &Peer{active: false}

	state.Active(address2)
	if assert.Len(t, state.peers, 2) {
		if assert.Contains(t, state.peers, address2) {
			p := state.peers[address2]
			assert.True(t, p.active)
		}
	}
}

func TestStateInactive(t *testing.T) {
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	state := &State{peers: make(map[string]*Peer)}

	err := state.Inactive(address1)
	if assert.NotNil(t, err) {
		assert.Equal(t, ErrNotExist, errors.Cause(err))
	}
	assert.Len(t, state.peers, 0)

	state.peers[address2] = &Peer{active: true}

	err = state.Inactive(address2)
	assert.Nil(t, err)
	if assert.Len(t, state.peers, 1) {
		if assert.Contains(t, state.peers, address2) {
			p := state.peers[address2]
			assert.False(t, p.active)
		}
	}
}

func TestStateReceived(t *testing.T) {

	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	hash1 := types.Hash{0x1}
	state := &State{peers: make(map[string]*Peer)}

	err := state.Received(address1, hash1)
	if assert.NotNil(t, err) {
		assert.Equal(t, ErrNotExist, errors.Cause(err))
	}

	state.peers[address2] = &Peer{yes: make(map[types.Hash]struct{})}
	err = state.Received(address2, hash1)
	assert.Nil(t, err)
	if assert.Len(t, state.peers[address2].yes, 1) {
		assert.Contains(t, state.peers[address2].yes, hash1)
	}
}

func TestStateSeen(t *testing.T) {

	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	address3 := "192.0.2.150:1337"

	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}

	state := &State{peers: make(map[string]*Peer)}

	peer1 := &Peer{yes: make(map[types.Hash]struct{})}
	peer2 := &Peer{yes: make(map[types.Hash]struct{})}

	peer1.yes[hash1] = struct{}{}
	peer1.yes[hash2] = struct{}{}

	state.peers[address1] = peer1
	state.peers[address2] = peer2

	seen, err := state.Seen(address1)
	assert.Nil(t, err)
	assert.ElementsMatch(t, []types.Hash{hash1, hash2}, seen)

	seen, err = state.Seen(address2)
	assert.Nil(t, err)
	assert.Equal(t, []types.Hash{}, seen)

	_, err = state.Seen(address3)
	if assert.NotNil(t, err) {
		assert.Equal(t, ErrNotExist, errors.Cause(err))
	}
}

func TestStateAddressesCount(t *testing.T) {

	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	address3 := "192.0.2.150:1337"
	address4 := "192.0.2.250:1337"

	hash1 := types.Hash{0x1}
	hash2 := types.Hash{0x2}

	vectors := map[string]struct {
		peers     map[string]*Peer
		filters   []FilterFunc
		addresses []string
	}{
		"active": {
			peers: map[string]*Peer{
				address1: &Peer{active: true},
				address2: &Peer{active: false},
				address3: &Peer{active: true},
				address4: &Peer{active: false},
			},
			filters: []FilterFunc{
				IsActive(true),
			},
			addresses: []string{
				address1,
				address3,
			},
		},
		"entity": {
			peers: map[string]*Peer{
				address1: &Peer{yes: map[types.Hash]struct{}{hash1: struct{}{}}},
				address2: &Peer{yes: map[types.Hash]struct{}{hash2: struct{}{}}},
				address3: &Peer{yes: map[types.Hash]struct{}{}},
				address4: &Peer{yes: map[types.Hash]struct{}{hash1: struct{}{}, hash2: struct{}{}}},
			},
			filters: []FilterFunc{
				HasEntity(EntityYes, hash2),
			},
			addresses: []string{
				address2,
				address4,
			},
		},
		"both": {
			peers: map[string]*Peer{
				address1: &Peer{
					active: true,
					yes:    map[types.Hash]struct{}{hash1: struct{}{}},
				},
				address2: &Peer{
					active: true,
					yes:    map[types.Hash]struct{}{},
				},
				address3: &Peer{
					active: false,
					yes:    map[types.Hash]struct{}{hash1: struct{}{}},
				},
				address4: &Peer{
					active: false,
					yes:    map[types.Hash]struct{}{},
				},
			},
			filters: []FilterFunc{
				IsActive(false),
				HasEntity(EntityYes, hash1),
			},
			addresses: []string{
				address3,
			},
		},
	}

	for name, vector := range vectors {
		state := &State{peers: vector.peers}
		addresses := state.Addresses(vector.filters...)
		count := state.Count(vector.filters...)
		assert.ElementsMatch(t, vector.addresses, addresses, name)
		assert.Equal(t, uint(len(vector.addresses)), count, name)
	}
}

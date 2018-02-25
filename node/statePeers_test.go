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
)

func TestNewState(t *testing.T) {
	state := newState()
	assert.NotNil(t, state.actives)
	assert.NotNil(t, state.tags)
}

func TestStateActive(t *testing.T) {
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	state := &simpleState{actives: make(map[string]bool)}

	state.Active(address1)
	if assert.Len(t, state.actives, 1) {
		assert.Contains(t, state.actives, address1)
	}

	state.Active(address1)
	assert.Len(t, state.actives, 1)

	state.Active(address2)
	if assert.Len(t, state.actives, 2) {
		assert.Contains(t, state.actives, address2)
	}
}

func TestStateInactive(t *testing.T) {
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	state := &simpleState{actives: make(map[string]bool)}

	state.actives[address1] = true
	state.actives[address2] = true
	state.Inactive(address1)
	if assert.Len(t, state.actives, 1) {
		assert.NotContains(t, state.actives, address1)
	}

	state.Inactive(address1)
	assert.Len(t, state.actives, 1)

	state.Inactive(address2)
	assert.Len(t, state.actives, 0)
}

func TestStateActives(t *testing.T) {
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	state := &simpleState{actives: make(map[string]bool)}

	actives := state.Actives()
	assert.Empty(t, actives)

	state.actives[address1] = true
	state.actives[address2] = true
	actives = state.Actives()
	assert.ElementsMatch(t, []string{address1, address2}, actives)
}

func TestStateTag(t *testing.T) {
	id1 := []byte{1, 2, 3, 4}
	id2 := []byte{5, 6, 7, 8}
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	state := &simpleState{tags: make(map[string][]string)}

	state.Tag(address1, id1)
	if assert.Len(t, state.tags[string(id1)], 1) {
		assert.Contains(t, state.tags[string(id1)], address1)
	}

	assert.Empty(t, state.tags[string(id2)])

	state.Tag(address1, id2)
	if assert.Len(t, state.tags[string(id2)], 1) {
		assert.Contains(t, state.tags[string(id2)], address1)
	}

	state.Tag(address2, id1)
	if assert.Len(t, state.tags[string(id1)], 2) {
		assert.Contains(t, state.tags[string(id1)], address2)
	}
}

func TestStateTags(t *testing.T) {
	id := []byte{1, 2, 3, 4}
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	state := &simpleState{tags: make(map[string][]string)}

	state.tags[string(id)] = []string{address1, address2}
	tags := state.Tags(id)
	assert.ElementsMatch(t, []string{address1, address2}, tags)
}

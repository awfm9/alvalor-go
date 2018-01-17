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

func TestIsAny(t *testing.T) {
	vectors := map[string]struct {
		filter   func(e *entry) bool
		entry    *entry
		expected bool
	}{
		"active entry": {
			filter:   isAny(),
			entry:    &entry{Active: true},
			expected: true,
		},
		"inactive entry": {
			filter:   isAny(),
			entry:    &entry{Active: false},
			expected: true,
		},
	}
	for name, vector := range vectors {
		actual := vector.filter(vector.entry)
		assert.Equalf(t, vector.expected, actual, "Is any filter wrong result for %v", name)
	}
}

func TestIsActive(t *testing.T) {
	vectors := map[string]struct {
		filter   func(e *entry) bool
		entry    *entry
		expected bool
	}{
		"active entry & active filter": {
			filter:   isActive(true),
			entry:    &entry{Active: true},
			expected: true,
		},
		"inactive entry & active filter": {
			filter:   isActive(true),
			entry:    &entry{Active: false},
			expected: false,
		},
		"active entry & inactive filter": {
			filter:   isActive(false),
			entry:    &entry{Active: true},
			expected: false,
		},
		"inactive entry & inactive filter": {
			filter:   isActive(false),
			entry:    &entry{Active: false},
			expected: true,
		},
	}
	for name, vector := range vectors {
		actual := vector.filter(vector.entry)
		assert.Equalf(t, vector.expected, actual, "Is active filter wrong result for %v", name)
	}
}

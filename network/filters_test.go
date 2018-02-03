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

func TestIsNot(t *testing.T) {
	vectors := map[string]struct {
		filter   func(string) bool
		entry    string
		expected bool
	}{
		"empty filter": {
			filter:   isNot(nil),
			entry:    "192.0.2.100:1337",
			expected: true,
		},
		"one entry filter positive": {
			filter:   isNot([]string{"192.0.2.100:1337"}),
			entry:    "192.0.2.100:1337",
			expected: false,
		},
		"one entry filter negative": {
			filter:   isNot([]string{"192.168.2.200:1337"}),
			entry:    "192.0.2.100:1337",
			expected: true,
		},
		"multi entry filter positive": {
			filter:   isNot([]string{"192.0.2.100:1337", "192.168.2.200:1337", "192.168.2.300:1337"}),
			entry:    "192.0.2.100:1337",
			expected: false,
		},
		"multi entry filter negative": {
			filter:   isNot([]string{"192.168.2.200:1337", "192.168.2.300:1337", "192.168.2.400:1337"}),
			entry:    "192.0.2.100:1337",
			expected: true,
		},
	}
	for name, vector := range vectors {
		actual := vector.filter(vector.entry)
		assert.Equalf(t, vector.expected, actual, "Is not filter wrong result for %v", name)
	}
}

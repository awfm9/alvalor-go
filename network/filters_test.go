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
	"time"

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

func TestIsScoreAbove(t *testing.T) {

	address := "192.0.1.100:1337"

	rep := &ReputationManagerMock{}
	rep.On("Score", address).Return(float64(10))

	vectors := map[string]struct {
		threshold float32
		expected  bool
	}{
		"above": {
			threshold: 9,
			expected:  true,
		},
		"equal": {
			threshold: 10,
			expected:  false,
		},
		"below": {
			threshold: 11,
			expected:  false,
		},
	}

	for name, vector := range vectors {
		filter := isScoreAbove(rep, vector.threshold)
		actual := filter(address)
		assert.Equalf(t, vector.expected, actual, "Is score above wrong result for %v", name)
	}
}

func TestIsLastBefore(t *testing.T) {

	address := "192.0.1.100:1337"
	now := time.Now()

	rep := &ReputationManagerMock{}
	rep.On("Fail", address).Return(now)

	vectors := map[string]struct {
		cutoff   time.Time
		expected bool
	}{
		"before": {
			cutoff:   now.Add(time.Second),
			expected: true,
		},
		"equal": {
			cutoff:   now,
			expected: false,
		},
		"after": {
			cutoff:   now.Add(-time.Second),
			expected: false,
		},
	}

	for name, vector := range vectors {
		filter := isFailBefore(rep, vector.cutoff)
		actual := filter(address)
		assert.Equalf(t, vector.expected, actual, "Is score above wrong result for %v", name)
	}
}

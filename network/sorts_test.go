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

func TestByRandom(t *testing.T) {
	e := &entry{}
	sort := byRandom()
	mismatch := false
	for i := 0; i < 100; i++ {
		ok1 := sort(e, e)
		ok2 := sort(e, e)
		if ok1 != ok2 {
			mismatch = true
			break
		}
	}
	assert.True(t, mismatch, "By random sort always returns same result")
}

func TestByScore(t *testing.T) {
	sort := byScore()
	vectors := map[string]struct {
		entry1   *entry
		entry2   *entry
		expected bool
	}{
		"first score higher": {
			entry1:   &entry{Success: 1, Failure: 0},
			entry2:   &entry{Success: 0, Failure: 1},
			expected: true,
		},
		"first score lower": {
			entry1:   &entry{Success: 0, Failure: 1},
			entry2:   &entry{Success: 1, Failure: 0},
			expected: false,
		},
		"both equal score": {
			entry1:   &entry{Success: 1, Failure: 1},
			entry2:   &entry{Success: 1, Failure: 1},
			expected: false,
		},
	}
	for name, vector := range vectors {
		actual := sort(vector.entry1, vector.entry2)
		assert.Equalf(t, vector.expected, actual, "By score sort wrong result for %v", name)
	}
}

func TestByScoreFunc(t *testing.T) {
	value := map[bool]float64{true: 1, false: 0}
	vectors := map[string]struct {
		entry1   *entry
		entry2   *entry
		score    func(*entry) float64
		expected bool
	}{
		"first active on active scoring": {
			entry1:   &entry{Active: true},
			entry2:   &entry{Active: false},
			score:    func(e *entry) float64 { return value[e.Active] },
			expected: true,
		},
		"second active on active scoring": {
			entry1:   &entry{Active: false},
			entry2:   &entry{Active: true},
			score:    func(e *entry) float64 { return value[e.Active] },
			expected: false,
		},
		"both active on active scoring": {
			entry1:   &entry{Active: true},
			entry2:   &entry{Active: true},
			score:    func(e *entry) float64 { return value[e.Active] },
			expected: false,
		},
		"first active on inactive scoring": {
			entry1:   &entry{Active: true},
			entry2:   &entry{Active: false},
			score:    func(e *entry) float64 { return value[!e.Active] },
			expected: false,
		},
		"second active on inactive scoring": {
			entry1:   &entry{Active: false},
			entry2:   &entry{Active: true},
			score:    func(e *entry) float64 { return value[!e.Active] },
			expected: true,
		},
		"both active on inactive scoring": {
			entry1:   &entry{Active: true},
			entry2:   &entry{Active: true},
			score:    func(e *entry) float64 { return value[!e.Active] },
			expected: false,
		},
	}
	for name, vector := range vectors {
		actual := byScoreFunc(vector.score)(vector.entry1, vector.entry2)
		assert.Equalf(t, vector.expected, actual, "By score func sort wrong result for %v", name)
	}
}

func TestByHashFunc(t *testing.T) {
}

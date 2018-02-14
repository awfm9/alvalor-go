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
	"crypto/md5"
	"crypto/sha256"
	"hash"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestByRandom(t *testing.T) {
	address := "192.0.2.100:1337"
	sort := byRandom()
	mismatch := false
	for i := 0; i < 100; i++ {
		ok1 := sort(address, address)
		ok2 := sort(address, address)
		if ok1 != ok2 {
			mismatch = true
			break
		}
	}
	assert.True(t, mismatch, "By random sort always returns same result")
}

func TestByScore(t *testing.T) {
	address1 := "192.0.2.100:1337"
	address2 := "192.0.2.200:1337"
	rep := newSimpleReputationManager()
	sort := byScore(rep)
	vectors := map[string]struct {
		score1   float32
		score2   float32
		expected bool
	}{
		"first score higher": {
			score1:   1,
			score2:   -1,
			expected: true,
		},
		"first score lower": {
			score1:   -1,
			score2:   1,
			expected: false,
		},
		"both equal score": {
			score1:   0,
			score2:   0,
			expected: false,
		},
	}
	for name, vector := range vectors {
		rep.scores[address1] = vector.score1
		rep.scores[address2] = vector.score2
		actual := sort(address1, address2)
		assert.Equalf(t, vector.expected, actual, "By reputation sort wrong result for %v", name)
	}
}

func TestByHashFunc(t *testing.T) {
	vectors := map[string]struct {
		address1 string
		address2 string
		hash     hash.Hash
		expected bool
	}{
		"same md5": {
			address1: "192.0.2.1:1234", // d0f88d6c87767262ba8e93d6acccd784
			address2: "192.0.2.1:1234", // d0f88d6c87767262ba8e93d6acccd784
			hash:     md5.New(),
			expected: false,
		},
		"first lower md5": {
			address1: "192.0.2.2:1234", // 7f83fddecaba901abfd469d899958433
			address2: "192.0.2.1:4321", // d0f88d6c87767262ba8e93d6acccd784
			hash:     md5.New(),
			expected: true,
		},
		"second lower md5": {
			address1: "192.0.2.1:1234", // d0f88d6c87767262ba8e93d6acccd784
			address2: "192.0.2.2:4321", // 7f83fddecaba901abfd469d899958433
			hash:     md5.New(),
			expected: false,
		},
		"same address sha256": {
			address1: "192.0.2.1:1234", // 37fcff24bf62035b2b08020afc08b4fecd4fcffce57ab23518e3561ff0fe76b9
			address2: "192.0.2.1:4321", // 37fcff24bf62035b2b08020afc08b4fecd4fcffce57ab23518e3561ff0fe76b9
			hash:     sha256.New(),
			expected: false,
		},
		"first lower sha256": {
			address1: "192.0.2.1:1234", // 37fcff24bf62035b2b08020afc08b4fecd4fcffce57ab23518e3561ff0fe76b9
			address2: "192.0.2.2:4321", // 9a6b293639db1e588add3900fe817a3ed3b9822a99e4799098e550a2d70b7e1f
			hash:     sha256.New(),
			expected: true,
		},
		"second lower sha256": {
			address1: "192.0.2.2:1234", // 9a6b293639db1e588add3900fe817a3ed3b9822a99e4799098e550a2d70b7e1f
			address2: "192.0.2.1:4321", // 37fcff24bf62035b2b08020afc08b4fecd4fcffce57ab23518e3561ff0fe76b9
			hash:     sha256.New(),
			expected: false,
		},
	}
	for name, vector := range vectors {
		actual := byIPHash(vector.hash)(vector.address1, vector.address2)
		assert.Equalf(t, vector.expected, actual, "By hash func sort wrong result for %v", name)
	}
}

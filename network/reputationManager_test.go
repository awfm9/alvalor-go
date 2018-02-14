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

func TestNewReputationManager(t *testing.T) {
	rep := newSimpleReputationManager()
	assert.NotNil(t, rep.scores)
	assert.NotNil(t, rep.fails)
}

func TestReputationManagerFailure(t *testing.T) {
	address := "192.0.2.100:1337"
	rep := simpleReputationManager{
		scores: make(map[string]float32),
		fails:  make(map[string]time.Time),
	}
	rep.Failure(address)
	assert.Equal(t, float32(-1), rep.scores[address])
	assert.WithinDuration(t, time.Now(), rep.fails[address], time.Second)
}

func TestReputationManagerSuccess(t *testing.T) {
	address := "192.0.2.100:1337"
	rep := simpleReputationManager{
		scores: make(map[string]float32),
	}
	rep.Success(address)
	assert.Equal(t, float32(1), rep.scores[address])
}

func TestReputationManagerScore(t *testing.T) {
	score := float32(13)
	address := "192.0.2.100:1337"
	rep := simpleReputationManager{
		scores: map[string]float32{address: score},
	}
	assert.Equal(t, score, rep.Score(address))
	assert.Equal(t, float32(0), rep.Score("whatever"))
}

func TestReputationManagerLast(t *testing.T) {
	last := time.Now()
	address := "192.0.2.100:1337"
	rep := simpleReputationManager{
		fails: map[string]time.Time{address: last},
	}
	assert.Equal(t, last, rep.Fail(address))
	assert.Equal(t, time.Time{}, rep.Fail("whatever"))
}

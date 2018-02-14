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

func TestNewReputationManager(t *testing.T) {
	rep := newSimpleReputationManager()
	assert.NotNil(t, rep.scores)
}

func TestReputationManagerFailure(t *testing.T) {
	score := float32(13)
	address := "192.0.2.100:1337"
	rep := simpleReputationManager{
		scores: map[string]float32{address: score},
	}
	rep.Failure(address)
	assert.Equal(t, score-1, rep.scores[address])
}

func TestReputationManagerSuccess(t *testing.T) {
	score := float32(13)
	address := "192.0.2.100:1337"
	rep := simpleReputationManager{
		scores: map[string]float32{address: score},
	}
	rep.Success(address)
	assert.Equal(t, score+1, rep.scores[address])
}

func TestReputationManagerScore(t *testing.T) {
	score := float32(13)
	address := "192.0.2.100:1337"
	rep := simpleReputationManager{
		scores: map[string]float32{address: score},
	}
	assert.Equal(t, score, rep.Score(address))
}

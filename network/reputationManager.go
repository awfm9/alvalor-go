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

import "sync"

type reputationManager interface {
	Failure(address string)
	Success(address string)
	Score(address string) float32
}

type simpleReputationManager struct {
	sync.Mutex
	scores map[string]float32
}

func newSimpleReputationManager() *simpleReputationManager {
	return &simpleReputationManager{
		scores: make(map[string]float32),
	}
}

func (rm *simpleReputationManager) Failure(address string) {
	rm.Lock()
	defer rm.Unlock()
	rm.scores[address]--
}

func (rm *simpleReputationManager) Success(address string) {
	rm.Lock()
	defer rm.Unlock()
	rm.scores[address]++
}

func (rm *simpleReputationManager) Score(address string) float32 {
	rm.Lock()
	defer rm.Unlock()
	return rm.scores[address]
}

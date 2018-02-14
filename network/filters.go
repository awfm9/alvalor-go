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

import "time"

func isNot(addresses []string) func(string) bool {
	lookup := make(map[string]struct{})
	for _, address := range addresses {
		lookup[address] = struct{}{}
	}
	return func(address string) bool {
		_, ok := lookup[address]
		return !ok
	}
}

func isScoreAbove(rep reputationManager, threshold float32) func(string) bool {
	return func(address string) bool {
		return rep.Score(address) > threshold
	}
}

func isFailBefore(rep reputationManager, cutoff time.Time) func(string) bool {
	return func(address string) bool {
		return rep.Fail(address).Before(cutoff)
	}
}

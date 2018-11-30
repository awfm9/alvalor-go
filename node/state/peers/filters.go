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

package peers

import "github.com/alvalor/alvalor-go/types"

// FilterFunc represens a filter that allows us to filter the list of
// returned peers from the peer states.
type FilterFunc func(*Peer) bool

// EntityHas describes how a peer knows an entity.
type EntityHas uint8

// List of possible state on entity knowledge.
const (
	EntityMaybe EntityHas = iota
	EntityYes
	EntityNo
)

// HasEntity allows us to select only peers that have or don't have a certain
// entity (such as transaction or header).
func HasEntity(has EntityHas, hash types.Hash) func(*Peer) bool {
	return func(p *Peer) bool {
		_, yes := p.yes[hash]
		_, no := p.no[hash]
		switch has {
		case EntityYes:
			return yes
		case EntityNo:
			return no
		default:
			return yes || !no
		}
	}
}

// IsActive allows us to only select peers that are currently active or
// inactive.
func IsActive(active bool) func(*Peer) bool {
	return func(p *Peer) bool {
		return p.active == active
	}
}

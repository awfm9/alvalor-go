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

package handler

import (
	"github.com/alvalor/alvalor-go/types"
)

// Paths is responsible for tracking the paths in our tree of headers and
// downloading the entities required for the best one.
type Paths interface {
	Follow(path []types.Hash) error
	Signal(hash types.Hash) error
}

// Downloads manages downloading of entities by keeping track of pending
// downloads and load balancing across available peers.
type Downloads interface {
	Start(hash types.Hash) error
	Cancel(hash types.Hash) error
}

// Events represents a manager for events for external subscribers.
type Events interface {
	Subscribe(sub chan<- interface{}, filters ...func(interface{}) bool)
	Unsubscribe(sub chan<- interface{})
	Header(hash types.Hash) error
	Transaction(hash types.Hash) error
}

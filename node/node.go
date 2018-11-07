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

package node

import (
	"sync"
)

// EventHandler represents a handler to process events. We could use a function, but
// using an interface makes mocking for tests easier.
type EventHandler interface {
	Process(interface{})
}

// Run will run the node package with the given event handler on the stream of
// input events.
func Run(wg *sync.WaitGroup, events <-chan interface{}, handler EventHandler) {
	wg.Add(1)
	defer wg.Done()
	for event := range events {
		handler.Process(event)
	}
}

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
	"testing"

	"github.com/alvalor/alvalor-go/network"
	"github.com/stretchr/testify/mock"
)

type HandlerMock struct {
	mock.Mock
}

func (hm *HandlerMock) Process(wg *sync.WaitGroup, event interface{}) {
	hm.Called(wg, event)
	wg.Done()
}

func TestRun(t *testing.T) {

	// create a dummy waitgroup
	wg := &sync.WaitGroup{}

	// create three events for testing
	e1 := network.Connected{Address: "192.0.2.1"}
	e2 := network.Disconnected{Address: "192.0.2.2"}
	e3 := network.Received{Address: "192.0.2.3"}

	// create a stream of three events in a closed channel
	events := make(chan interface{}, 5)
	events <- e1
	events <- e2
	events <- e3
	close(events)

	// create the mock handler instance
	handler := &HandlerMock{}
	handler.On("Process", mock.Anything, mock.Anything)

	// run the node on the input channel
	wg.Add(1)
	Run(wg, events, handler)
	wg.Wait()

	// assert the handler was called with all three events
	if handler.AssertNumberOfCalls(t, "Process", 3) {
		handler.AssertCalled(t, "Process", wg, e1)
		handler.AssertCalled(t, "Process", wg, e2)
		handler.AssertCalled(t, "Process", wg, e3)
	}
}

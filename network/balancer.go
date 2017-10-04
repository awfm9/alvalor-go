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

// Balancer is responsible for balancing the number of peers at certain
// intervals.
type Balancer struct {
	events   chan<- interface{}
	minPeers int
	maxPeers int
	interval time.Duration
	sig      chan struct{}
}

// NewBalancer creates a new balancer in charge of sending balancing requests to
// the node.
func NewBalancer(events chan<- interface{}, options ...func(*Balancer)) *Balancer {
	bal := &Balancer{
		events:   events,
		minPeers: 8,
		maxPeers: 16,
		interval: time.Second,
		sig:      make(chan struct{}),
	}
	for _, option := range options {
		option(bal)
	}
	return bal
}

// Start will start the balancing impulses.
func (bal *Balancer) Start() {
	ticker := time.NewTicker(bal.interval)
Loop:
	for {
		select {
		case <-bal.sig:
			break Loop
		case <-ticker.C:
			bal.events <- Balance{Min: bal.minPeers, Max: bal.maxPeers}
		}
	}
	ticker.Stop()
	close(bal.sig)
}

// Close will end the balancer.
func (bal *Balancer) Close() {
	bal.sig <- struct{}{}
	<-bal.sig
}

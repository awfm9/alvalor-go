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

import "go.uber.org/zap"

// Processor represents a manager for events, executing the respective actions we
// want depending on the events.
type Processor struct {
	log    *zap.Logger
	events <-chan interface{}
}

// NewProcessor creates a new manager of network events.
func NewProcessor(log *zap.Logger, events <-chan interface{}) *Processor {
	pro := &Processor{
		log:    log,
		events: events,
	}
	return pro
}

// Process will launch the processing of the processor.
func (pro *Processor) Process() {
	for event := range pro.events {
		switch e := event.(type) {
		case Connection:
			_ = e
		case Disconnection:
		case Failure:
		case Violation:
		case Message:
		default:
		}
	}
}

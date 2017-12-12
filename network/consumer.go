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

// Consumer represents an outside consumer of the package and provides the
// public API.
type Consumer struct {
	events chan<- interface{}
}

// NewConsumer creates a new consumer using the given channels to perform
// actions for the external user on the networking companents.
func NewConsumer(log *zap.Logger, events chan<- interface{}) *Consumer {
	con := &Consumer{}
	return con
}

// Peers will return a list of all connected peers.
func (con *Consumer) Peers() ([]string, error) {
	return nil, nil
}

// Send will send the given message to the given peers.
func (con *Consumer) Send(addresses []string, message interface{}) error {
	return nil
}

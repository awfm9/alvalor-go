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
	"time"

	"go.uber.org/zap"
)

// DefaultConfig represents the default configuration of a node.
var DefaultConfig = Config{
	log:        DefaultLogger,
	book:       DefaultBook,
	codec:      DefaultCodec,
	subscriber: nil,
	network:    Odin,
	server:     false,
	address:    "",
	minPeers:   3,
	maxPeers:   10,
	balance:    time.Millisecond * 100,
	heartbeat:  time.Second * 1,
	timeout:    time.Second * 3,
	discovery:  time.Second * 2,
}

// DefaultLogger represents the default logger we use for our node.
var DefaultLogger, _ = zap.NewDevelopment()

// DefaultCodec represents the default codec we use for our serialization.
var DefaultCodec = SimpleCodec{}

// Config represents the configuration parameters available to configure a node on the peer-to-peer
// network.
type Config struct {
	log        *zap.Logger
	book       Book
	codec      Codec
	subscriber chan<- interface{}
	network    []byte
	server     bool
	address    string
	minPeers   uint
	maxPeers   uint
	balance    time.Duration
	heartbeat  time.Duration
	timeout    time.Duration
	discovery  time.Duration
}

// SetLog is a configuration function that allows us to change the logger instance.
func SetLog(log *zap.Logger) func(*Config) {
	return func(cfg *Config) {
		cfg.log = log
	}
}

// SetBook is a configuration function that allows us to change the address book implementation.
func SetBook(book Book) func(*Config) {
	return func(cfg *Config) {
		cfg.book = book
	}
}

// SetCodec is a configuration function that allows us to change the serialization codec.
func SetCodec(codec Codec) func(*Config) {
	return func(cfg *Config) {
		cfg.codec = codec
	}
}

// SetSubscriber is a configuration function that lets us define a subscriber to listen to all
// network messages & events not handled internally by the network library.
func SetSubscriber(sub chan<- interface{}) func(*Config) {
	return func(cfg *Config) {
		cfg.subscriber = sub
	}
}

// SetBalance allows us to set the frequency at which we will check whether we are above or below
// the defined peer thresholds.
func SetBalance(balance time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.balance = balance
	}
}

// SetHeartbeat allows us to set the frequency at which we will send heartbeats to inactive peers.
func SetHeartbeat(heartbeat time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.heartbeat = heartbeat
	}
}

// SetTimeout allows us to define the timeout at which we consider a peer unresponsive and drop it.
func SetTimeout(timeout time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.timeout = timeout
	}
}

// SetDiscovery allows us to define how often we want to launch discovery requests on the network to
// find additional peer addresses.
func SetDiscovery(discovery time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.discovery = discovery
	}
}

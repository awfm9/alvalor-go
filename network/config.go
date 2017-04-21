// Copyright (c) 2017 The Veltor Authors
//
// This file is part of Veltor.
//
// Veltor is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Veltor is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Veltor.  If not, see <http://www.gnu.org/licenses/>.

package network

import (
	"time"
	"go.uber.org/zap"
)

// DefaultConfig variable.
var DefaultConfig = Config{
	log:        defaultLogger,
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

var defaultLogger, _ = zap.NewDevelopment()

// Config struct.
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

// SetLog function.
func SetLog(log *zap.Logger) func(*Config) {
	return func(cfg *Config) {
		cfg.log = log
	}
}

// SetBook function.
func SetBook(book Book) func(*Config) {
	return func(cfg *Config) {
		cfg.book = book
	}
}

// SetCodec function.
func SetCodec(codec Codec) func(*Config) {
	return func(cfg *Config) {
		cfg.codec = codec
	}
}

// SetSubscriber function.
func SetSubscriber(sub chan<- interface{}) func(*Config) {
	return func(cfg *Config) {
		cfg.subscriber = sub
	}
}

// SetNetwork function.
func SetNetwork(net []byte) func(*Config) {
	return func(cfg *Config) {
		cfg.network = net
	}
}

// SetServer function.
func SetServer(server bool) func(*Config) {
	return func(cfg *Config) {
		cfg.server = server
	}
}

// SetAddress function.
func SetAddress(address string) func(*Config) {
	return func(cfg *Config) {
		cfg.address = address
	}
}

// SetMinPeers function.
func SetMinPeers(minPeers uint) func(*Config) {
	return func(cfg *Config) {
		cfg.minPeers = minPeers
	}
}

// SetMaxPeers function.
func SetMaxPeers(maxPeers uint) func(*Config) {
	return func(cfg *Config) {
		cfg.maxPeers = maxPeers
	}
}

// SetBalance function.
func SetBalance(balance time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.balance = balance
	}
}

// SetHeartbeat function.
func SetHeartbeat(heartbeat time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.heartbeat = heartbeat
	}
}

// SetTimeout function.
func SetTimeout(timeout time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.timeout = timeout
	}
}

// SetDiscovery function.
func SetDiscovery(discovery time.Duration) func(*Config) {
	return func(cfg *Config) {
		cfg.discovery = discovery
	}
}

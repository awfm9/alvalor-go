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

// Config represents the configuration parameters available to configure a node
// on the peer-to-peer network.
type Config struct {
	network  []byte
	nonce    []byte
	listen   bool
	address  string
	minPeers uint
	maxPeers uint
	interval time.Duration
}

// SetNetwork allows us to configure a custom network ID.
func SetNetwork(network []byte) func(*Config) {
	return func(cfg *Config) {
		cfg.network = network
	}
}

// SetListen allows us to configure a custom server status.
func SetListen(listen bool) func(*Config) {
	return func(cfg *Config) {
		cfg.listen = listen
	}
}

// SetAddress allows us to configure a custom listen address.
func SetAddress(address string) func(*Config) {
	return func(cfg *Config) {
		cfg.address = address
	}
}

// SetMinPeers allows us to configure a custom number for minimum peers.
func SetMinPeers(minPeers uint) func(*Config) {
	return func(cfg *Config) {
		cfg.minPeers = minPeers
	}
}

// SetMaxPeers allows us to configure a custom number for maximum peers.
func SetMaxPeers(maxPeers uint) func(*Config) {
	return func(cfg *Config) {
		cfg.maxPeers = maxPeers
	}
}

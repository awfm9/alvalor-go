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

// Peer contains all the information related to a peer.
type Peer struct {
	Address string
}

// Registry represents a registry for all peers with their current states.
type Registry struct {
	peers   map[string]Peer
	pending uint
}

// NewRegistry creates a new initialized peer registry.
func NewRegistry() *Registry {
	return &Registry{
		peers: make(map[string]Peer),
	}
}

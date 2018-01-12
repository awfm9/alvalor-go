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

package miner

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/alvalor/alvalor-go/types"
	"golang.org/x/crypto/blake2b"
)

// Naive represents a naive miner which is not efficient but easy to understand.
type Naive struct {
	header types.Header
	parent chan []byte
	state  chan []byte
	delta  chan []byte
	target chan []byte
	out    chan types.Header
	stop   chan struct{}
}

// NewNaive creates a new naive miner initialized with the given parameters.
func NewNaive(parent []byte, state []byte, delta []byte, miner []byte, target []byte) *Naive {
	header := types.Header{
		Parent: parent,
		State:  state,
		Delta:  delta,
		Miner:  miner,
		Target: target,
		Nonce:  0,
		Time:   time.Now(),
	}
	return &Naive{
		header: header,
		parent: make(chan []byte, 1),
		state:  make(chan []byte, 1),
		delta:  make(chan []byte, 1),
		target: make(chan []byte, 1),
		stop:   make(chan struct{}),
	}
}

// Start will start the mining process.
func (nv *Naive) Start() <-chan types.Header {
	nv.out = make(chan types.Header)
	go nv.mine()
	return nv.out
}

// Stop will stop the mining process.
func (nv *Naive) Stop() {
	close(nv.out)
	close(nv.stop)
}

// Parent will update the block we try to mine with a new parent hash.
func (nv *Naive) Parent(parent []byte) {
	nv.parent <- parent
}

// Delta will update the block we try to mine with a new transaction root hash.
func (nv *Naive) Delta(delta []byte) {
	nv.delta <- delta
}

// Target will update the target difficulty we are trying to mine for.
func (nv *Naive) Target(target []byte) {
	nv.target <- target
}

// mine is the mining loop.
func (nv *Naive) mine() {
Loop:
	for {
		select {
		case <-nv.stop:
			break Loop
		case parent := <-nv.parent:
			nv.header.Parent = parent
			nv.header.Nonce = 0
			continue Loop
		case state := <-nv.state:
			nv.header.State = state
			nv.header.Nonce = 0
			continue Loop
		case delta := <-nv.delta:
			nv.header.Delta = delta
			nv.header.Nonce = 0
		case target := <-nv.target:
			nv.header.Target = target
			nv.header.Nonce = 0
		default:
			// go on
		}
		nv.header.Time = time.Now()
		h, _ := blake2b.New256(nil)
		_, _ = h.Write(nv.header.Parent)
		_, _ = h.Write(nv.header.State)
		_, _ = h.Write(nv.header.Delta)
		_, _ = h.Write(nv.header.Miner)
		_, _ = h.Write(nv.header.Target)
		ts := make([]byte, 8)
		binary.LittleEndian.PutUint64(ts, uint64(nv.header.Time.Unix()))
		_, _ = h.Write(ts)
		hash := h.Sum(nil)
		if bytes.Compare(hash, nv.header.Target) < 0 {
			nv.out <- nv.header
		}
		nv.header.Nonce++
	}
}

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
	"golang.org/x/crypto/sha3"
)

// Naive represents a naive miner which is not efficient but easy to understand.
type Naive struct {
	data   []byte
	parent chan types.Header
	delta  chan []byte
	target chan []byte
	stop   chan struct{}
	out    chan types.Header
}

// NewNaive creates a new naive miner initialized with the given parameters.
func NewNaive(miner []byte, parent types.Header, delta []byte, target []byte) *Naive {
	nv := &Naive{
		data:   make([]byte, 176),
		parent: make(chan types.Header, 1),
		delta:  make(chan []byte),
		target: make(chan []byte),
		stop:   make(chan struct{}),
	}
	copy(nv.data[0:32], parent.Hash())
	copy(nv.data[32:64], parent.State)
	copy(nv.data[64:96], delta)
	copy(nv.data[96:128], miner)
	copy(nv.data[128:160], target)
	return nv
}

// Start will start the mining process.
func (nv *Naive) Start() <-chan types.Header {
	nv.out = make(chan types.Header)
	go nv.mine()
	return nv.out
}

// Stop will stop the mining process.
func (nv *Naive) Stop() {
	close(nv.stop)
	<-nv.out
}

// Parent will update the block we try to mine with a new parent hash.
func (nv *Naive) Parent(parent types.Header) {
	nv.parent <- parent
}

// Delta will update the block we try to mine with a new transaction root hash.
func (nv *Naive) Delta(delta []byte) {
	nv.delta <- delta
}

// Target will change the target difficulty of the block hash.
func (nv *Naive) Target(target []byte) {
	nv.target <- target
}

// mine is the mining loop.
func (nv *Naive) mine() {
	nonce := uint64(0)
	ticker := time.NewTicker(time.Second)
Loop:
	for {
		select {
		case <-nv.stop:
			break Loop
		case parent := <-nv.parent:
			copy(nv.data[0:32], parent.Hash())
			copy(nv.data[32:64], parent.State)
			nonce = 0
			continue Loop
		case delta := <-nv.delta:
			copy(nv.data[64:96], delta)
			nonce = 0
			continue Loop
		case ts := <-ticker.C:
			binary.LittleEndian.PutUint64(nv.data[160:168], uint64(ts.Unix()))
			nonce = 0
		default:
			// go on
		}
		binary.LittleEndian.PutUint64(nv.data[168:176], nonce)
		hash := sha3.Sum256(nv.data)
		if bytes.Compare(hash[:], nv.data[128:160]) < 0 {
			unix := binary.LittleEndian.Uint64(nv.data[160:168])
			header := types.Header{
				Parent: make([]byte, 32),
				State:  make([]byte, 32),
				Delta:  make([]byte, 32),
				Miner:  make([]byte, 32),
				Target: make([]byte, 32),
				Time:   time.Unix(int64(unix), 0),
				Nonce:  nonce,
			}
			copy(header.Parent, nv.data[0:32])
			copy(header.State, nv.data[32:64])
			copy(header.Delta, nv.data[64:96])
			copy(header.Miner, nv.data[96:128])
			copy(header.Target, nv.data[128:160])
			nv.out <- header
			nonce = 0
			continue
		}
		nonce++
	}
	close(nv.out)
}

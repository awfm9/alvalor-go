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
	"errors"

	"github.com/alvalor/alvalor-go/types"
)

type chainManager interface {
	Height() (uint32, error)
	BestHash() ([]byte, error)
	HashByHeight(height uint32) ([]byte, error)
}

type simpleChain struct {
	best   *types.Block
	blocks map[string]*types.Block
}

func newChain() *simpleChain {
	return &simpleChain{}
}

func (b *simpleChain) Add(block *types.Block) error {
	id := string(block.ID())
	_, ok := b.blocks[id]
	if ok {
		return errors.New("block already known")
	}
	parent := string(block.Parent)
	_, ok = b.blocks[parent]
	if !ok {
		return errors.New("block parent unknown")
	}
	b.blocks[id] = block
	if block.Height > b.best.Height {
		b.best = block
	}
	return nil
}

func (b *simpleChain) Height() (uint32, error) {
	return b.best.Height, nil
}

func (b *simpleChain) BestHash() ([]byte, error) {
	return b.best.ID(), nil
}

func (b *simpleChain) HashByHeight(height uint32) ([]byte, error) {
	if height > b.best.Height {
		return nil, errors.New("invalid block height")

	}
	for _, block := range b.blocks {
		if block.Height == height {
			return block.ID(), nil
		}
	}
	return nil, errors.New("unknown block height")
}

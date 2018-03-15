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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/alvalor/alvalor-go/types"
)

func TestNewPool(t *testing.T) {
	codec := &CodecMock{}
	store := &StoreMock{}
	pool := newPool(codec, store)
	assert.Equal(t, codec, pool.codec)
	assert.Equal(t, store, pool.store)
}

func TestPoolAdd(t *testing.T) {

	tx1 := &types.Transaction{Data: []byte{1, 2, 3, 4}}
	tx2 := &types.Transaction{Data: []byte{4, 5, 6, 7}}
	tx3 := &types.Transaction{Data: []byte{8, 9, 0, 1}}

	codec := &CodecMock{}
	codec.On("Encode", mock.Anything, tx1).Return(nil)
	codec.On("Encode", mock.Anything, tx2).Return(errors.New("could not encode"))
	codec.On("Encode", mock.Anything, tx3).Return(nil)

	store := &StoreMock{}
	store.On("Put", tx1.Hash(), mock.Anything).Return(nil)
	store.On("Put", tx2.Hash(), mock.Anything).Return(nil)
	store.On("Put", tx3.Hash(), mock.Anything).Return(errors.New("could not put"))

	pool := simplePool{
		codec:  codec,
		store:  store,
		hashes: make(map[types.Hash]struct{}),
	}

	err := pool.Add(tx1)
	assert.Nil(t, err)

	err = pool.Add(tx2)
	assert.NotNil(t, err)

	err = pool.Add(tx3)
	assert.NotNil(t, err)
}

func TestPoolGet(t *testing.T) {

	id1 := [32]byte{1, 2, 3, 4}
	id2 := [32]byte{4, 5, 6, 7}
	id3 := [32]byte{8, 9, 0, 1}

	tx1 := &types.Transaction{Data: id1[:]}
	tx3 := &types.Transaction{Data: id3[:]}

	store := &StoreMock{}
	store.On("Get", id1).Return(id1, nil)
	store.On("Get", id2).Return(id2, errors.New("could not get"))
	store.On("Get", id3).Return(id3, nil)

	// this is not ideal, but the way returns are evaluated immediately, we can't
	// use the closure of the Run function to switch on the buffer contents
	codec := &CodecMock{}
	codec.On("Decode", mock.Anything).Return(tx1, nil).Once()
	codec.On("Decode", mock.Anything).Return(tx3, errors.New("could not decode"))

	pool := simplePool{codec: codec, store: store}

	tx, err := pool.Get(id1)
	assert.Nil(t, err)
	assert.Equal(t, tx1, tx)

	_, err = pool.Get(id2)
	assert.NotNil(t, err)

	_, err = pool.Get(id3)
	assert.NotNil(t, err)
}

func TestPoolRemove(t *testing.T) {

	id1 := [32]byte{1, 2, 3, 4}
	id2 := [32]byte{4, 5, 6, 7}

	store := &StoreMock{}
	store.On("Del", id1).Return(nil)
	store.On("Del", id2).Return(errors.New("could not del"))

	pool := simplePool{
		store:  store,
		hashes: make(map[types.Hash]struct{}),
	}

	err := pool.Remove(id1)
	assert.Nil(t, err)

	err = pool.Remove(id2)
	assert.NotNil(t, err)
}

func TestPoolKnown(t *testing.T) {

	id1 := [32]byte{1, 2, 3, 4}
	id2 := [32]byte{5, 6, 7, 8}

	pool := simplePool{
		hashes: make(map[types.Hash]struct{}),
	}
	pool.hashes[id1] = struct{}{}

	ok := pool.Known(id1)
	assert.True(t, ok)

	ok = pool.Known(id2)
	assert.False(t, ok)
}

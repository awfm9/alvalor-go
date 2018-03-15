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

package wallet

import (
	"bytes"

	"github.com/pkg/errors"
	argon2 "github.com/tvdburgt/go-argon2"
	"golang.org/x/crypto/ed25519"
)

// Salt variable.
var salt = []byte{
	0x53, 0x78, 0x3e, 0x4c,
	0x94, 0x78, 0x59, 0x18,
	0x8a, 0x9b, 0x31, 0xe7,
	0x4d, 0xed, 0x1d, 0x29,
}

// Store represents the key store for a user wallet with a main key used as the root of all key
// derivations and an argon2 context to parameterize the key derivation algorithm.
type Store struct {
	root []byte
	ctx  argon2.Context
}

// NewStore initializes a new key store for a wallet from the given seed. The seed should ideally
// be 256-bit long and will be used to derive the root key of the key store. The preconfigured
// argon2 context uses 3 iterations, 1 gigabyte of memory, 4 lanes and an output length of 96 bytes.
func NewStore(seed []byte) (*Store, error) {
	s := &Store{
		ctx: argon2.Context{
			Iterations:  3,
			Memory:      1 << 16,
			Parallelism: 4,
			HashLen:     96,
			Mode:        argon2.ModeArgon2i,
			Version:     argon2.Version13,
		},
	}
	root, err := s.generate(seed, salt)
	if err != nil {
		return nil, errors.Wrap(err, "could not generate root key")
	}
	s.root = root
	return s, nil
}

// generate uses some input data and a salt to derive and generate a new futhark private key.
// First, we create 96 bytes of data by running the data and salt through the argon2 key derivation
// hash function. We then feed the first 64 bytes of this hash into the futhark key generation
// function to get our key, while appending the last 32 bytes to the final key as the salt for the
// derivations starting at this key.
func (s *Store) generate(data []byte, salt []byte) ([]byte, error) {
	hash, err := argon2.Hash(&s.ctx, data, salt)
	if err != nil {
		return nil, errors.Wrap(err, "could not compute hash")
	}
	_, priv, err := ed25519.GenerateKey(bytes.NewReader(hash[:64]))
	if err != nil {
		return nil, errors.Wrap(err, "could not generate private key")
	}
	key := append([]byte(priv), hash[64:]...)
	return key, nil
}

// derive takes a parent key and a derivation index to get the child key of the input key at the
// given index. The index is given as a byte array to reduce additional copying of data. It uses
// the first 64 bytes of the parent key concatenated with the index as the input data for the key
// generation, while using the last 32 bytes of the key as the salt.
func (s *Store) derive(key []byte, index []byte) ([]byte, error) {
	data := append(key[:64], index...)
	key, err := s.generate(data, key[64:])
	if err != nil {
		return nil, errors.Wrap(err, "could not derive level one key")
	}
	return key, nil
}

// Key takes a list of derivation levels, where the length of the slice represents the total depth
// we want to derive to, while each byte value corresponds to the index we want to derive to on
// the corresponding level.
func (s *Store) Key(levels [][]byte) ([]byte, error) {
	var err error
	key := s.root
	for i, level := range levels {
		key, err = s.derive(key, level)
		if err != nil {
			return nil, errors.Wrapf(err, "could not derive level %v key", i)
		}
	}
	return key, nil
}

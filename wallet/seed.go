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

package wallet

import (
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"

	"github.com/andreburgaud/crypt2go/ecb"
	"github.com/pkg/errors"
	bip39 "github.com/tyler-smith/go-bip39"
)

// NewSeed function.
func NewSeed() ([]byte, error) {
	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		return nil, errors.Wrap(err, "could not get random data")
	}
	return seed, nil
}

// SeedFromMnemonic function.
func SeedFromMnemonic(mnemonic string) ([]byte, error) {
	seed, err := bip39.MnemonicToByteArray(mnemonic)
	if err != nil {
		return nil, errors.Wrap(err, "could not create seed from mnemonic")
	}
	return seed, nil
}

// SeedToMnemonic function.
func SeedToMnemonic(seed []byte) (string, error) {
	mnemonic, err := bip39.NewMnemonic(seed)
	if err != nil {
		return "", errors.Wrap(err, "could not create mnemonic from seed")
	}
	return mnemonic, nil
}

// SeedFromEncrypted function.
func SeedFromEncrypted(encrypted []byte, password string) ([]byte, error) {
	hash := sha256.Sum256([]byte(password))
	cipher, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize cipher")
	}
	mode := ecb.NewECBDecrypter(cipher)
	seed := make([]byte, len(encrypted))
	mode.CryptBlocks(seed, encrypted)
	return seed, nil
}

// SeedToEncrypted function.
func SeedToEncrypted(seed []byte, password string) ([]byte, error) {
	hash := sha256.Sum256([]byte(password))
	cipher, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, errors.Wrap(err, "could not initialize cipher")
	}
	mode := ecb.NewECBEncrypter(cipher)
	encrypted := make([]byte, len(seed))
	mode.CryptBlocks(encrypted, seed)
	return encrypted, nil
}

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
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"

	"github.com/andreburgaud/crypt2go/ecb"
	"github.com/pkg/errors"
	bip39 "github.com/tyler-smith/go-bip39"
)

// NewSeed creates a new 256-bit random seed to be used for wallet key store initialization.
func NewSeed() ([]byte, error) {
	seed := make([]byte, 32)
	_, err := rand.Read(seed)
	if err != nil {
		return nil, errors.Wrap(err, "could not get random data")
	}
	return seed, nil
}

// SeedFromMnemonic takes a mnemonic as defined in BIP39 and turns it back into a byte slice used
// as seed for a wallet key store.
func SeedFromMnemonic(mnemonic string) ([]byte, error) {
	seed, err := bip39.MnemonicToByteArray(mnemonic)
	if err != nil {
		return nil, errors.Wrap(err, "could not create seed from mnemonic")
	}
	return seed, nil
}

// SeedToMnemonic takes a random byte slice representing the seed of a wallet key store and turns it
// into the BIP39 mnemonic representation of words to be used as a offline backup for user wallets.
func SeedToMnemonic(seed []byte) (string, error) {
	mnemonic, err := bip39.NewMnemonic(seed)
	if err != nil {
		return "", errors.Wrap(err, "could not create mnemonic from seed")
	}
	return mnemonic, nil
}

// SeedFromEncrypted takes a AES ECB mode encrypted seed and a password and will decode the original
// seed from the input. The seed should not be more than 256 bits long to avoid issues with the ECB
// mode application of AES.
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

// SeedToEncrypted takes a random wallet seed and a password and uses AES ECB mode encryption to
// get the output that can be written to disk securely. The seed shouldn't be more than 256 bits to
// avoid ECB mode security concerns.
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

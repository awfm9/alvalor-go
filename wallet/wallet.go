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

// Package wallet implements functions to support portable and flexible user wallets to manage and
// generate keys. It includes a key store which uses deterministic hierarchical key derivation to
// create the same groups and series of keys for a specific seed, as well as a number of helper
// functions to write the seed to disk, read it from disk and back it up offline using a mnemonic.
package wallet

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

type pathManager interface {
	Add(header *types.Header) error
	Has(hash []byte) bool
	BestHash() []byte
}

type simplePath struct {
	best    *types.Header
	headers map[string]*types.Header
}

func newPath() *simplePath {
	return &simplePath{
		headers: make(map[string]*types.Header),
	}
}

func (s *simplePath) Add(header *types.Header) error {
	hash := string(header.Hash())
	_, ok := s.headers[hash]
	if ok {
		return errors.New("header already known")
	}
	parent := string(header.Parent)
	_, ok = s.headers[parent]
	if !ok {
		return errors.New("parent not known")
	}
	s.headers[hash] = header
	if header.Height > s.best.Height {
		s.best = header
	}
	return nil
}

func (s *simplePath) Has(hash []byte) bool {
	_, ok := s.headers[string(hash)]
	return ok
}

func (s *simplePath) BestHash() []byte {
	return s.best.Hash()
}

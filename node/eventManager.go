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
	"time"
)

type eventManager interface {
	Transaction(hash types.Hash) error
}

type simpleEventManager struct {
	subscriber chan<- interface{}
}

func (mgr *simpleEventManager) Transaction(hash types.Hash) error {
	select {
	case mgr.subscriber <- Transaction{hash: hash}:
	case <-time.After(10 * time.Millisecond):
		return errors.New("subscriber stalling")
	}
	return nil
}

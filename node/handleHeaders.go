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
	"github.com/Workiva/go-datastructures/queue"
	"github.com/alvalor/alvalor-go/types"
	"github.com/rs/zerolog"
	"sync"
	"time"
)

func handleHeaders(log zerolog.Logger, wg *sync.WaitGroup, net Network, receivers map[string]*queue.Queue, stop <-chan struct{}) {
	defer wg.Done()

	log = log.With().Str("component", "headers").Logger()

	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case <-stop:
			ticker.Stop()
			return
		case <-ticker.C:
		}

		for addr, queue := range receivers {
			if queue.Len() > 0 {
				msg, err := queue.Get(1)

				if err != nil {
					continue
				}

				if len(msg) == 0 {
					continue
				}

				header := msg[0].(types.Header)

				err = net.Send(addr, header)
				if err != nil {
					log.Error().Err(err).Msg("could not send header")
					continue
				}
			}
		}
	}
}

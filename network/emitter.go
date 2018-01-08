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

package network

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Handler is used to interface with the manager.
type Handler interface {
	StartHandlers()
}

func handleEmitting(log zerolog.Logger, wg *sync.WaitGroup, cfg *Config, mgr Handler, stop <-chan struct{}) {
	defer wg.Done()

	// extract configuration parameters we care about
	var (
		interval = cfg.interval
	)

	// set up logging with start/stop messages
	log = log.With().Str("component", "emitter").Dur("interval", interval).Logger()
	log.Info().Msg("emitting routine started")
	defer log.Info().Msg("emitting routine stopped")

	// on each ticker, execute handlers function until we quit
	ticker := time.NewTicker(interval)
Loop:
	for {
		select {
		case <-stop:
			break Loop
		case <-ticker.C:
			mgr.StartHandlers()
		}
	}
	ticker.Stop()
}

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

package event

//
// func handleStream(log zerolog.Logger, wg *sync.WaitGroup, stream <-chan interface{}, subscribers <-chan *subscriber) {
// 	defer wg.Done()
//
// 	log = log.With().Str("component", "stream").Logger()
// 	log.Debug().Msg("subscriber routine started")
// 	defer log.Debug().Msg("subscriber routine stopped")
//
// 	var subs []*subscriber
//
// Loop:
// 	for {
// 		select {
//
// 		case sub, ok := <-subscribers:
// 			if !ok {
// 				break Loop
// 			}
// 			subs = append(subs, sub)
//
// 		case event, ok := <-stream:
// 			if !ok {
// 				break Loop
// 			}
// 			for _, sub := range subs {
// 				for _, filter := range sub.filters {
// 					if filter(event) {
// 						continue Loop
// 					}
// 				}
// 				select {
// 				case sub.channel <- event:
// 				case <-time.After(10 * time.Millisecond):
// 					log.Warn().Msg("subscriber is stalling")
// 				}
// 			}
// 		}
// 	}
// }

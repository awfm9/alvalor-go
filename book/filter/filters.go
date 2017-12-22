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

package filter

import "github.com/alvalor/alvalor-go/book"

// Active represents filter to select active/inactive entries in Sample method
func Active(active bool) func(e *book.Entry) bool {
	return func(e *book.Entry) bool {
		return e.Active == active
	}
}

// Any reperesents filter to select any entries in Sample method
func Any() func(e *book.Entry) bool {
	return func(e *book.Entry) bool {
		return true
	}
}

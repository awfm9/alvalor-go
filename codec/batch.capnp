# Copyright (c) 2017 The Alvalor Authors
#
# This file is part of Alvalor.
#
# Alvalor is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# Alvalor is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

using Go = import "/go.capnp";
$Go.package("codec");
$Go.import("codec");

using Transaction = import "transaction.capnp".Transaction;

@0xe60ceb912269b107;
struct Batch {
  transactions @0 :List(Transaction);
}

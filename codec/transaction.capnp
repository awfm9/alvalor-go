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

using Transfer = import "transfer.capnp".Transfer;
using Fee = import "fee.capnp".Fee;

@0xb5f3d18a6c743283;
struct Transaction {
  transfers @0 :List(Transfer);
  fees @1: List(Fee);
  data @2: Data;
  nonce @3: UInt64;
  signatures @4: List(Data);
}

using Go = import "/go.capnp";
$Go.package("proto");
$Go.import("proto");

using Ping = import "ping.capnp".Ping;
using Pong = import "pong.capnp".Pong;
using Discover = import "discover.capnp".Discover;
using Peers = import "peers.capnp".Peers;

@0x904d4f3f728c7f04;
struct Z {
	union {
		ping @0 :Ping;
		pong @1 :Pong;
		discover @2: Discover;
		peers @3: Peers;
		text @4: Text;
		data @5: Data;
	}
}

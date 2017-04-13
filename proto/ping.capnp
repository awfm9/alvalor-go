using Go = import "/go.capnp";
$Go.package("proto");
$Go.import("proto");

@0xc4411f23835fd9b0;
struct Ping {
	nonce @0 :UInt32;
}

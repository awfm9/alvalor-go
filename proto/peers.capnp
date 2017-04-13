using Go = import "/go.capnp";
$Go.package("proto");
$Go.import("proto");

@0xb8fb51aaf7fc2d2f;
struct Peers {
	addresses @0 :List(Text);
}

// Code generated by capnpc-go. DO NOT EDIT.

package codec

import (
	capnp "zombiezen.com/go/capnproto2"
	text "zombiezen.com/go/capnproto2/encoding/text"
	schemas "zombiezen.com/go/capnproto2/schemas"
)

type Transaction struct{ capnp.Struct }

// Transaction_TypeID is the unique identifier for the type Transaction.
const Transaction_TypeID = 0xec9fd906d129035f

func NewTransaction(s *capnp.Segment) (Transaction, error) {
	st, err := capnp.NewStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 4})
	return Transaction{st}, err
}

func NewRootTransaction(s *capnp.Segment) (Transaction, error) {
	st, err := capnp.NewRootStruct(s, capnp.ObjectSize{DataSize: 0, PointerCount: 4})
	return Transaction{st}, err
}

func ReadRootTransaction(msg *capnp.Message) (Transaction, error) {
	root, err := msg.RootPtr()
	return Transaction{root.Struct()}, err
}

func (s Transaction) String() string {
	str, _ := text.Marshal(0xec9fd906d129035f, s.Struct)
	return str
}

func (s Transaction) Transfers() (Transfer_List, error) {
	p, err := s.Struct.Ptr(0)
	return Transfer_List{List: p.List()}, err
}

func (s Transaction) HasTransfers() bool {
	p, err := s.Struct.Ptr(0)
	return p.IsValid() || err != nil
}

func (s Transaction) SetTransfers(v Transfer_List) error {
	return s.Struct.SetPtr(0, v.List.ToPtr())
}

// NewTransfers sets the transfers field to a newly
// allocated Transfer_List, preferring placement in s's segment.
func (s Transaction) NewTransfers(n int32) (Transfer_List, error) {
	l, err := NewTransfer_List(s.Struct.Segment(), n)
	if err != nil {
		return Transfer_List{}, err
	}
	err = s.Struct.SetPtr(0, l.List.ToPtr())
	return l, err
}

func (s Transaction) Fees() (Fee_List, error) {
	p, err := s.Struct.Ptr(1)
	return Fee_List{List: p.List()}, err
}

func (s Transaction) HasFees() bool {
	p, err := s.Struct.Ptr(1)
	return p.IsValid() || err != nil
}

func (s Transaction) SetFees(v Fee_List) error {
	return s.Struct.SetPtr(1, v.List.ToPtr())
}

// NewFees sets the fees field to a newly
// allocated Fee_List, preferring placement in s's segment.
func (s Transaction) NewFees(n int32) (Fee_List, error) {
	l, err := NewFee_List(s.Struct.Segment(), n)
	if err != nil {
		return Fee_List{}, err
	}
	err = s.Struct.SetPtr(1, l.List.ToPtr())
	return l, err
}

func (s Transaction) Data() ([]byte, error) {
	p, err := s.Struct.Ptr(2)
	return []byte(p.Data()), err
}

func (s Transaction) HasData() bool {
	p, err := s.Struct.Ptr(2)
	return p.IsValid() || err != nil
}

func (s Transaction) SetData(v []byte) error {
	return s.Struct.SetData(2, v)
}

func (s Transaction) Signatures() (capnp.DataList, error) {
	p, err := s.Struct.Ptr(3)
	return capnp.DataList{List: p.List()}, err
}

func (s Transaction) HasSignatures() bool {
	p, err := s.Struct.Ptr(3)
	return p.IsValid() || err != nil
}

func (s Transaction) SetSignatures(v capnp.DataList) error {
	return s.Struct.SetPtr(3, v.List.ToPtr())
}

// NewSignatures sets the signatures field to a newly
// allocated capnp.DataList, preferring placement in s's segment.
func (s Transaction) NewSignatures(n int32) (capnp.DataList, error) {
	l, err := capnp.NewDataList(s.Struct.Segment(), n)
	if err != nil {
		return capnp.DataList{}, err
	}
	err = s.Struct.SetPtr(3, l.List.ToPtr())
	return l, err
}

// Transaction_List is a list of Transaction.
type Transaction_List struct{ capnp.List }

// NewTransaction creates a new list of Transaction.
func NewTransaction_List(s *capnp.Segment, sz int32) (Transaction_List, error) {
	l, err := capnp.NewCompositeList(s, capnp.ObjectSize{DataSize: 0, PointerCount: 4}, sz)
	return Transaction_List{l}, err
}

func (s Transaction_List) At(i int) Transaction { return Transaction{s.List.Struct(i)} }

func (s Transaction_List) Set(i int, v Transaction) error { return s.List.SetStruct(i, v.Struct) }

func (s Transaction_List) String() string {
	str, _ := text.MarshalList(0xec9fd906d129035f, s.List)
	return str
}

// Transaction_Promise is a wrapper for a Transaction promised by a client call.
type Transaction_Promise struct{ *capnp.Pipeline }

func (p Transaction_Promise) Struct() (Transaction, error) {
	s, err := p.Pipeline.Struct()
	return Transaction{s}, err
}

const schema_b5f3d18a6c743283 = "x\xdaL\xcd1K\xf3P\x14\xc6\xf1\xe7\xb9'\xed\xfb" +
	":\xb4\xf6\xda\xac\xc1Y\x07Q\xc1\xa5 \xe8\xa6[\xaf" +
	"8\xb9\xe8%MK@\xd2\x92\\\x87\x82\x9b\x8b\x0aB" +
	"\x05\xf7\xce\x01WG?\x82CG?\x81\x08N\xba{" +
	"%\x15\xd1\xed\xf0\xfbsx\xd6\x07\xdcQ\x1b\xb5\xe9\x7f" +
	"\xc0\xf4ju\x7f,+\xb3\xfa\xf3\xf4\x0dz\x89\xfeb" +
	"\xd3\x9d^\xcf>\x1eP\x0b\xfe\x01\xed\xad\xe0\xbd\xbd;" +
	"\xbf\xb6\x83\x17x,x\x97\xdb\xac\xb0\xb1S\xe90[" +
	"\x8b\xed(\x1bu\x0e\xe7\xb4\x1c\xbbt\x98uI\xd3\x92" +
	"\x00\x08\x08h{\x00\x98\x13\xa19W\xd4d\xc8\x0a\xc7" +
	"\xab\x80qB3Q\xd4J\x85T\x80\xbe\xa9\xf0Rh" +
	"\xee\x14\xb5HH\x01\xf4\xed\x11`&Bs\xaf\xf8=" +
	"\xdcOr\xb0`\x13\xec\x0a\xd9\xf2et\xb5\xd7\xd9\x7f" +
	"\xfc\x04X\xe1b?I\xfe\xd4q\xeb5*\xa3\xf2\xe9" +
	"\xa7\xf6\xac\xb3l@\xb1\x01\xfa\"\x1dd\xd6\x9d\xe5\x90" +
	"\xdf\x97\xaa5\xc1\xaf\x00\x00\x00\xff\xff\x87\xa1A\xd0"

func init() {
	schemas.Register(schema_b5f3d18a6c743283,
		0xec9fd906d129035f)
}

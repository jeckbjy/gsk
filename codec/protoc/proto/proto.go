// proto protobuf编解码的简单实现,限制是仅仅支持工具生成协议,好处是不依赖官方库,而且代码量也少
package proto

import "errors"

var ErrNotSupport = errors.New("protobuf not support")

// Text is implemented by generated protocol buffer messages.
type Message interface {
	Reset()
	String() string
	ProtoMessage()
}

// Marshaler is the interface representing objects that can marshal themselves.
type Marshaler interface {
	Marshal() ([]byte, error)
}

// newMarshaler is the interface representing objects that can marshal themselves.
//
// This exists to support protoc-gen-go generated messages.
// The proto package will stop type-asserting to this interface in the future.
//
// DO NOT DEPEND ON THIS.
type newMarshaler interface {
	XXX_Size() int
	XXX_Marshal(b []byte, deterministic bool) ([]byte, error)
}

// Marshal takes a protocol buffer message
// and encodes it into the wire format, returning the data.
// This is the main entry point.
func Marshal(pb Message) ([]byte, error) {
	if m, ok := pb.(newMarshaler); ok {
		siz := m.XXX_Size()
		b := make([]byte, 0, siz)
		return m.XXX_Marshal(b, false)
	}
	if m, ok := pb.(Marshaler); ok {
		// If the message can marshal itself, let it do it, for compatibility.
		// NOTE: This is not efficient.
		return m.Marshal()
	}

	return nil, ErrNotSupport
	// in case somehow we didn't generate the wrapper
	// if pb == nil {
	// 	return nil, ErrNil
	// }

	// var info InternalMessageInfo
	// siz := info.Size(pb)
	// b := make([]byte, 0, siz)
	// return info.Marshal(b, pb, false)
}

// Unmarshaler is the interface representing objects that can
// unmarshal themselves.  The argument points to data that may be
// overwritten, so implementations should not keep references to the
// buffer.
// Unmarshal implementations should not clear the receiver.
// Any unmarshaled data should be merged into the receiver.
// Callers of Unmarshal that do not want to retain existing data
// should Reset the receiver before calling Unmarshal.
type Unmarshaler interface {
	Unmarshal([]byte) error
}

// newUnmarshaler is the interface representing objects that can
// unmarshal themselves. The semantics are identical to Unmarshaler.
//
// This exists to support protoc-gen-go generated messages.
// The proto package will stop type-asserting to this interface in the future.
//
// DO NOT DEPEND ON THIS.
type newUnmarshaler interface {
	XXX_Unmarshal([]byte) error
}

// Unmarshal parses the protocol buffer representation in buf and places the
// decoded result in pb.  If the struct underlying pb does not match
// the data in buf, the results can be unpredictable.
//
// Unmarshal resets pb before starting to unmarshal, so any
// existing data in pb is always removed. Use UnmarshalMerge
// to preserve and append to existing data.
func Unmarshal(buf []byte, pb Message) error {
	pb.Reset()
	if u, ok := pb.(newUnmarshaler); ok {
		return u.XXX_Unmarshal(buf)
	}
	if u, ok := pb.(Unmarshaler); ok {
		return u.Unmarshal(buf)
	}

	return ErrNotSupport
	// return NewBuffer(buf).Unmarshal(pb)
}

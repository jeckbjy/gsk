package encoder

type Encoder interface {
	Encode(v interface{}) ([]byte, error)
	Decode([]byte, interface{}) error
	String() string
}

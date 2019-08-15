package json

import (
	"encoding/json"

	"github.com/jeckbjy/gsk/util/encoder"
)

func New() encoder.Encoder {
	return &jencoder{}
}

type jencoder struct {
}

func (*jencoder) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (*jencoder) Decode(d []byte, v interface{}) error {
	return json.Unmarshal(d, v)
}

func (*jencoder) String() string {
	return "json"
}

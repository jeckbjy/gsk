package xml

import "encoding/xml"

type xencoder struct {
}

func (*xencoder) Encode(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

func (*xencoder) Decode(d []byte, v interface{}) error {
	return xml.Unmarshal(d, v)
}

func (*xencoder) String() string {
	return "xml"
}

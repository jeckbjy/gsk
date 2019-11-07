package crypto

import (
	"errors"
	"net/url"
)

// Encode encodes the values into ``URL encoded'' form
// ("bar=baz&foo=quux") sorted by key.
func EncodeValues(data interface{}) (string, error) {
	var uv url.Values
	switch d := data.(type) {
	case string:
		return d, nil
	case map[string]string:
		uv = url.Values{}
		for k, v := range d {
			uv.Add(k, v)
		}
	case url.Values:
		uv = d
	default:
		return "", errors.New("not support")
	}

	return uv.Encode(), nil
}

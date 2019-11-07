package xml

import (
	"encoding/xml"
	"io"
)

// 扩展xml数据结构,用于map和xml互转,只能用于根节点是xml且只有1个层级的数据
// 微信支付使用的是xml作为通信协议
type StringMap map[string]string

func (m StringMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name.Local = "xml"
	tokens := []xml.Token{start}

	for key, value := range m {
		t := xml.StartElement{Name: xml.Name{"", key}}
		tokens = append(tokens, t, xml.CharData(value), xml.EndElement{t.Name})
	}

	tokens = append(tokens, xml.EndElement{start.Name})

	for _, t := range tokens {
		err := e.EncodeToken(t)
		if err != nil {
			return err
		}
	}

	// flush to ensure tokens are written
	err := e.Flush()
	if err != nil {
		return err
	}

	return nil
}

type xmlMapEntry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

func (m StringMap) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		var e xmlMapEntry

		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		m[e.XMLName.Local] = e.Value
	}
	return nil
}

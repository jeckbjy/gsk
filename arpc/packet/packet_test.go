package packet

import (
	"io"
	"testing"

	"github.com/jeckbjy/gsk/codec/jsonc"
)

func TestEncode(t *testing.T) {
	body := map[string]string{
		"a": "a",
		"b": "b",
	}
	codec := jsonc.New()

	pkg := New()
	pkg.SetAck(true)
	pkg.SetMsgID(1)
	pkg.SetSeqID(10)
	pkg.SetName("test")
	pkg.SetService("game")
	pkg.SetCodec(codec)
	pkg.SetBody(body)
	if err := pkg.Encode(); err != nil {
		t.Fatal(err)
	}

	buf := pkg.Buffer()
	_, _ = buf.Seek(0, io.SeekStart)

	pkgd := New()
	pkgd.SetBuffer(buf)
	pkgd.SetCodec(codec)
	if err := pkgd.Decode(); err != nil {
		t.Fatal(err)
	}

	bodyd := make(map[string]string)

	if err := pkgd.DecodeBody(&bodyd); err != nil {
		t.Fatal(err)
	}
	t.Log(pkgd)
	t.Log(bodyd)
}

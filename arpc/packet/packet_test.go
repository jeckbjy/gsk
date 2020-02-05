package packet

import (
	"io"
	"testing"

	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/codec/jsonc"
	"github.com/jeckbjy/gsk/util/buffer"
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
	buf := buffer.New()
	if err := pkg.Encode(buf); err != nil {
		t.Fatal(err)
	}

	str := buf.String()
	t.Log(str)
	_, _ = buf.Seek(0, io.SeekStart)

	pkgd := New()
	pkgd.SetCodec(codec)
	if err := pkgd.Decode(buf); err != nil {
		t.Fatal(err)
	}

	bodyd := make(map[string]string)

	if err := arpc.DecodeBody(pkgd, &bodyd); err != nil {
		t.Fatal(err)
	}
	t.Log(pkgd)
	t.Log(bodyd)
}

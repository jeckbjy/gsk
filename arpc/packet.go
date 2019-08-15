package arpc

import (
	"fmt"
	"io"
	"strings"

	"github.com/jeckbjy/gsk/arpc/codec"
	"github.com/jeckbjy/gsk/util/buffer"
)

func NewPacket() IPacket {
	return &Packet{}
}

const (
	flagReply   = 0x01
	flagSeqID   = 0x02
	flagID      = 0x04
	flagName    = 0x08
	flagMethod  = 0x10
	flagService = 0x20
	flagError   = 0x40
	flagHead    = 0x80
)

func hasFlag(flags int, f int) bool {
	return (flags & f) != 0
}

// Packet 实现IPacket接口
// 编码方式:tag|[SeqID]|[ID]|[Name]|[Method]|[Service]|[Error]|[Head]|[Body]
// 数值使用UVarint编码
// String使用Len+Value编码
type Packet struct {
	seqID   string
	reply   bool
	id      int
	name    string
	method  string
	service string
	head    map[string]string
	body    interface{}
	err     string
	codec   codec.ICodec
	bytes   *buffer.Buffer
}

func (p *Packet) SeqID() string {
	return p.seqID
}

func (p *Packet) Reply() bool {
	return p.reply
}

func (p *Packet) ID() int {
	return p.id
}

func (p *Packet) Name() string {
	return p.name
}

func (p *Packet) Method() string {
	return p.method
}

func (p *Packet) Service() string {
	return p.service
}

func (p *Packet) Head() map[string]string {
	return p.head
}

func (p *Packet) Body() interface{} {
	return p.body
}

func (p *Packet) Error() string {
	return p.err
}

func (p *Packet) Codec() codec.ICodec {
	return p.codec
}

func (p *Packet) Bytes() *buffer.Buffer {
	return p.bytes
}

func (p *Packet) Value(key string) string {
	if p.head == nil {
		return ""
	}

	return p.head[key]
}

func (p *Packet) SetSeqID(seqid string) {
	p.seqID = seqid
}

func (p *Packet) SetReply(r bool) {
	p.reply = r
}

func (p *Packet) SetID(id int) {
	p.id = id
}

func (p *Packet) SetName(name string) {
	p.name = name
}

func (p *Packet) SetMethod(method string) {
	p.method = method
}

func (p *Packet) SetService(service string) {
	p.service = service
}

func (p *Packet) SetBody(body interface{}) {
	p.body = body
}

func (p *Packet) SetError(err string) {
	p.err = err
}

func (p *Packet) SetCodec(c codec.ICodec) {
	p.codec = c
}

func (p *Packet) SetBytes(b *buffer.Buffer) {
	p.bytes = b
}

func (p *Packet) SetHead(h map[string]string) {
	p.head = h
}

func (p *Packet) SetValue(key string, value string) {
	if p.head == nil {
		p.head = make(map[string]string)
	}
	p.head[key] = value
}

func (p *Packet) ParseBody(msg interface{}) error {
	if err := p.codec.Decode(p.bytes, msg); err != nil {
		return err
	}

	p.body = msg
	return nil
}

func (p *Packet) Encode(b *buffer.Buffer) error {
	pos := b.Pos()

	// 消息头一般不会很大
	w := buffer.Writer{}
	w.Init(b, 128)
	flag := uint8(0)
	w.PutByte(flag)
	if p.reply {
		flag |= flagReply
	}
	if w.PutLenString(p.seqID) {
		flag |= flagSeqID
	}
	if p.id > 0 {
		w.PutUVarint(uint(p.id))
		flag |= flagID
	}
	if w.PutLenString(p.name) {
		flag |= flagName
	}
	if w.PutLenString(p.method) {
		flag |= flagMethod
	}
	if w.PutLenString(p.service) {
		flag |= flagService
	}
	if w.PutLenString(p.err) {
		flag |= flagError
	}
	// head
	if len(p.head) > 0 {
		flag |= flagHead
		w.PutVarintLen(len(p.head))
		for k, v := range p.head {
			s := fmt.Sprintf("%s:%v", k, v)
			w.PutVarintLen(len(s))
			if len(s) > 0 {
				w.PutString(s)
			}
		}
	}

	// body
	var body *buffer.Buffer
	if buff, ok := p.body.(*buffer.Buffer); ok {
		body = buff
	} else if p.codec != nil {
		if err := p.codec.Encode(b, p.body); err != nil {
			return err
		}
	}
	if body != nil {
		b.AppendBuffer(body)
	}

	w.Flush()

	_, _ = b.Seek(int64(pos), io.SeekStart)
	_ = b.WriteByte(flag)
	_, _ = b.Seek(0, io.SeekEnd)

	return nil
}

func (p *Packet) Decode(b *buffer.Buffer) error {
	p.bytes = b

	r := buffer.Reader{}
	r.Init(b)

	f, err := b.ReadByte()
	if err != nil {
		return err
	}
	flags := int(f)
	if hasFlag(flags, flagReply) {
		p.reply = true
	}

	if hasFlag(flags, flagSeqID) {
		if p.seqID, err = r.ReadLenString(); err != nil {
			return err
		}
	}

	if hasFlag(flags, flagID) {
		if p.id, err = r.ReadVarintLen(); err != nil {
			return err
		}
	}

	if hasFlag(flags, flagName) {
		if p.name, err = r.ReadLenString(); err != nil {
			return err
		}
	}

	if hasFlag(flags, flagMethod) {
		if p.method, err = r.ReadLenString(); err != nil {
			return err
		}
	}

	if hasFlag(flags, flagService) {
		if p.method, err = r.ReadLenString(); err != nil {
			return err
		}
	}

	if hasFlag(flags, flagError) {
		if p.err, err = r.ReadLenString(); err != nil {
			return err
		}
	}

	if hasFlag(flags, flagHead) {
		count, err := r.ReadVarintLen()
		if err != nil {
			return err
		}

		for i := 0; i < count; i++ {
			str, err := r.ReadLenString()
			if err != nil {
				return err
			}
			if len(str) == 0 {
				continue
			}
			tokens := strings.SplitN(str, ":", 2)
			if len(tokens) == 2 {
				p.SetValue(tokens[0], tokens[1])
			}
		}
	}

	return nil
}

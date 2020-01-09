package packet

import (
	"errors"

	"github.com/jeckbjy/gsk/util/errorx"

	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/util/buffer"
)

func New() arpc.Packet {
	return &Packet{}
}

type Packet struct {
	contentType uint
	ack         bool
	status      arpc.Status
	seqID       string
	msgID       int
	name        string
	method      string
	service     string
	extras      []string
	heads       map[string]string
	body        interface{}
	buffer      *buffer.Buffer
	codec       codec.Codec
	internal    interface{} // 以下字段不需要序列化
}

func (p *Packet) Reset() {
	*p = Packet{}
}

func (p *Packet) IsAck() bool {
	return p.ack
}

func (p *Packet) SetAck(ack bool) {
	p.ack = ack
}

func (p *Packet) Code() int {
	return p.status.Code
}

func (p *Packet) Status() string {
	return p.status.Info
}

func (p *Packet) SetStatus(code int, info string) {
	p.status.Code = code
	p.status.Info = info
}

func (p *Packet) ContentType() int {
	return int(p.contentType)
}

func (p *Packet) SetContentType(ct int) {
	p.contentType = uint(ct)
}

func (p *Packet) SeqID() string {
	return p.seqID
}

func (p *Packet) SetSeqID(id string) {
	p.seqID = id
}

func (p *Packet) MsgID() int {
	return p.msgID
}

func (p *Packet) SetMsgID(id int) {
	p.msgID = id
}

func (p *Packet) Name() string {
	return p.name
}

func (p *Packet) SetName(name string) {
	p.name = name
}

func (p *Packet) Method() string {
	return p.method
}

func (p *Packet) SetMethod(m string) {
	p.method = m
}

func (p *Packet) Service() string {
	return p.service
}

func (p *Packet) SetService(service string) {
	p.service = service
}

func (p *Packet) Extra(key uint) string {
	if key < uint(len(p.extras)) {
		return p.extras[key]
	}

	return ""
}

func (p *Packet) SetExtra(key uint, value string) error {
	if key >= arpc.HFExtraMax {
		return errors.New("bad extra key, must be in [0-6]")
	}

	if key >= uint(len(p.extras)) {
		extras := make([]string, key+1)
		copy(extras, p.extras)
	}
	p.extras[key] = value

	return nil
}

func (p *Packet) Head(key string) string {
	if p.heads != nil {
		return p.heads[key]
	}

	return ""
}

func (p *Packet) SetHead(key string, value string) {
	if p.heads == nil {
		p.heads = make(map[string]string)
	}

	p.heads[key] = value
}

func (p *Packet) Body() interface{} {
	return p.body
}

func (p *Packet) SetBody(body interface{}) {
	p.body = body
}

func (p *Packet) Codec() codec.Codec {
	if p.codec == nil {
		p.codec = codec.Get(p.ContentType())
	}
	return p.codec
}

func (p *Packet) SetCodec(c codec.Codec) {
	p.codec = c
}

func (p *Packet) Buffer() *buffer.Buffer {
	return p.buffer
}

func (p *Packet) SetBuffer(b *buffer.Buffer) {
	p.buffer = b
}

func (p *Packet) Internal() interface{} {
	return p.internal
}

func (p *Packet) SetInternal(value interface{}) {
	p.internal = value
}

func (p *Packet) encodeNameMethod() string {
	if p.method != "" {
		return "/" + p.method
	} else {
		return p.name
	}
}

func (p *Packet) decodeNameMethod(str string) {
	if len(str) == 0 {
		return
	}
	if str[0] == '/' {
		p.method = str[1:]
	} else {
		p.name = str
	}
}

func (p *Packet) Encode(data *buffer.Buffer) error {
	if data == nil {
		return errorx.ErrInvalidParam
	}
	p.buffer = data

	// encode head
	w := Writer{}
	w.Init()
	w.WriteBool(p.ack, 1<<arpc.HFAck)
	w.WriteString(p.status.Encode(), 1<<arpc.HFStatus)
	w.WriteUint(p.contentType, 1<<arpc.HFContentType)
	w.WriteString(p.seqID, 1<<arpc.HFSeqID)
	w.WriteInt(p.msgID, 1<<arpc.HFMsgID)
	w.WriteString(p.encodeNameMethod(), 1<<arpc.HFNameMethod)
	w.WriteString(p.service, 1<<arpc.HFService)
	w.WriteMap(p.heads, 1<<arpc.HFHeadMap)
	// extras
	if len(p.extras) > 0 {
		for i, v := range p.extras {
			w.WriteString(v, 1<<(uint(i)+arpc.HFExtra))
		}
	}

	p.buffer.Append(w.Flush())

	// encode body
	if p.body != nil {
		if body, ok := p.body.(*buffer.Buffer); ok {
			// 已经序列化好
			p.buffer.AppendBuffer(body)
		} else {
			if p.codec == nil {
				p.codec = codec.Get(int(p.contentType))
			}
			if p.codec == nil {
				return errors.New("not found codec")
			}
			buf := buffer.New()
			if err := p.codec.Encode(buf, p.body); err != nil {
				return err
			}
			p.buffer.AppendBuffer(buf)
		}
	}

	return nil
}

func (p *Packet) Decode(data *buffer.Buffer) (err error) {
	if data == nil {
		return errorx.ErrInvalidParam
	}
	p.buffer = data
	r := Reader{}
	if err := r.Init(p.buffer); err != nil {
		return err
	}

	status := ""
	nameMethod := ""
	// decode head
	r.ReadBool(&p.ack, 1<<arpc.HFAck)
	if err := r.ReadString(&status, 1<<arpc.HFStatus); err != nil {
		return err
	}
	if err := r.ReadUint(&p.contentType, 1<<arpc.HFContentType); err != nil {
		return err
	}
	p.status.Decode(status)

	if err := r.ReadString(&p.seqID, 1<<arpc.HFSeqID); err != nil {
		return err
	}
	if err := r.ReadInt(&p.msgID, 1<<arpc.HFMsgID); err != nil {
		return err
	}
	if err := r.ReadString(&nameMethod, 1<<arpc.HFNameMethod); err != nil {
		return err
	}
	p.decodeNameMethod(nameMethod)
	if err := r.ReadString(&p.service, 1<<arpc.HFService); err != nil {
		return err
	}
	if err := r.ReadMap(&p.heads, 1<<arpc.HFHeadMap); err != nil {
		return err
	}

	// extra
	if r.HasFlag(uint64(arpc.HFExtraMask)) {
		for i := arpc.HFExtra; i < arpc.HFMax; i++ {
			s, err := r.ReadStringDirect()
			if err != nil {
				return err
			}
			if s == "" {
				continue
			}
			key := i - arpc.HFExtra
			_ = p.SetExtra(uint(key), s)
		}
	}

	p.codec = codec.Get(p.ContentType())

	if p.body != nil && p.codec != nil {
		if err := p.codec.Decode(p.buffer, p.body); err != nil {
			return err
		}
	}

	return nil
}

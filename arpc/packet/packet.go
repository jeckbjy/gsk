package packet

import (
	"errors"

	"github.com/jeckbjy/gsk/arpc"
	"github.com/jeckbjy/gsk/codec"
	"github.com/jeckbjy/gsk/util/buffer"
)

func New() arpc.Packet {
	return &Packet{}
}

const (
	FlagStatus  = 1 << 0
	FlagReply   = 1 << 1
	FlagID      = 1 << 2
	FlagName    = 1 << 3
	FlagSeqID   = 1 << 4
	FlagMethod  = 1 << 5
	FlagService = 1 << 6
	FlagHead    = 1 << 7
)

type Packet struct {
	reply   bool
	status  uint
	id      uint
	name    string
	seqId   string
	method  string
	service string
	head    map[string]string
	body    interface{}
	data    *buffer.Buffer
	codec   codec.Codec
}

func (p *Packet) Reply() bool {
	return p.reply
}

func (p *Packet) Status() uint {
	return p.status
}

func (p *Packet) ID() uint {
	return p.id
}

func (p *Packet) Name() string {
	return p.name
}

func (p *Packet) SeqID() string {
	return p.seqId
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

func (p *Packet) Data() *buffer.Buffer {
	return p.data
}

func (p *Packet) Value(key string) string {
	if p.head != nil {
		return p.head[key]
	}

	return ""
}

func (p *Packet) Codec() codec.Codec {
	return p.codec
}

func (p *Packet) SetStatus(v uint) {
	p.status = v
}

func (p *Packet) SetReply(v bool) {
	p.reply = v
}

func (p *Packet) SetID(id uint) {
	p.id = id
}

func (p *Packet) SetName(name string) {
	p.name = name
}

func (p *Packet) SetSeqID(seqId string) {
	p.seqId = seqId
}

func (p *Packet) SetMethod(m string) {
	p.method = m
}

func (p *Packet) SetService(s string) {
	p.service = s
}

func (p *Packet) SetHead(head map[string]string) {
	p.head = head
}

func (p *Packet) SetBody(body interface{}) {
	p.body = body
}

func (p *Packet) SetData(data *buffer.Buffer) {
	p.data = data
}

func (p *Packet) SetValue(key string, value string) {
	if p.head == nil {
		p.head = make(map[string]string)
	}
	p.head[key] = value
}

func (p *Packet) SetCodec(c codec.Codec) {
	p.codec = c
}

func (p *Packet) Parse(msg interface{}) error {
	if p.codec == nil {
		return errors.New("no codec")
	}

	if err := p.codec.Decode(p.data, msg); err != nil {
		return err
	}

	p.body = msg
	return nil
}

func (p *Packet) Encode(b *buffer.Buffer) error {
	data := buffer.Buffer{}
	w := Writer{}
	w.Init(&data, 128)
	w.WriteUint(p.status, FlagStatus)
	w.WriteBool(p.reply, FlagReply)
	w.WriteUint(p.id, FlagID)
	w.WriteString(p.name, FlagName)
	w.WriteString(p.seqId, FlagSeqID)
	w.WriteString(p.method, FlagMethod)
	w.WriteString(p.service, FlagService)
	w.WriteMap(p.head, FlagHead)
	w.Flush()
	b.AppendBuffer(&data)
	return nil
}

func (p *Packet) Decode(b *buffer.Buffer) error {
	r := Reader{}
	if err := r.Init(b); err != nil {
		return err
	}
	r.ReadUint(&p.status, FlagStatus)

	return nil
}

package redeem

import (
	"errors"
	"math"
	"math/bits"
	"math/rand"
	"time"
)

var (
	ErrBadChecksum = errors.New("bad checksum")
)

func New() *Generator {
	return &Generator{abc: NewAlphabet(DefaultABC, 1)}
}

func NewEx(abc *Alphabet, randBits uint, injectBits uint) *Generator {
	if randBits+injectBits != 0 {
		rand.Seed(time.Now().UnixNano())
	}
	if abc == nil {
		abc = NewAlphabet(DefaultABC, 1)
	}

	return &Generator{abc: abc, randBits: uint8(randBits), injectBits: uint8(injectBits)}
}

// 生成规则与限制
// 通过一个表格ID和自增序列生成唯一ID
// 要求表格ID大于0且不能超过65535,自增ID大于0且不超过2^30
// 默认没有随机
// 情况下通常占用5-9位字符,例如
// 1000000 -> URSVUQHEL
// 1000001 -> JYVTQZGWG
// 1000002 -> UMVUFLHWS
// 1000003 -> JYVTPHEDB
// 1000004 -> URSDGHEGL
// 1000005 -> JKHZFUCVU
// 1000006 -> URSDHAUSK
// 1000007 -> URSVUVTPB
// 1000008 -> UMAGWTZAX
// 1000009 -> UMAGLHWAL
type Generator struct {
	abc        *Alphabet
	randBits   uint8 // 尾部随机字节数
	injectBits uint8 // 注入随机字节数
}

func (g *Generator) Build(tabID, start, count uint) []string {
	results := make([]string, 0, count)
	for i := start; i < start+count; i++ {
		d := g.Generate(uint64(tabID), uint64(i))
		results = append(results, d)
	}

	return results
}

// 验证ID是否合法,并返回隐含的TableID信息
func (g *Generator) Verify(id string) (uint64, error) {
	x, err := g.abc.Decode(id)
	if err != nil {
		return 0, nil
	}

	return g.Decode(x)
}

func (g *Generator) Generate(tabId, counter uint64) string {
	x := g.Encode(tabId, counter)
	return g.abc.Encode(x)
}

// 编码格式:counter(32bit) + tabID(4-16bit) + flag(2bit) + random(4) + checksum(2bit)
func (g *Generator) Encode(tabID, counter uint64) uint64 {
	num := uint64(math.Ceil(float64(bits.Len64(tabID)) / 4))
	rid := counter<<(num*4+2) | tabID<<2 | uint64(num-1)

	if g.randBits > 0 {
		r := uint64(rand.Intn(1 << g.randBits))
		rid = rid<<g.randBits | r
	}

	if g.injectBits > 0 {
		rid = injectBits(rid, uint(g.injectBits))
	}

	// 添加两位校验码
	one := uint64(bits.OnesCount64(rid) % 4)
	result := rid<<2 | one
	return result
}

func (g *Generator) Decode(id uint64) (uint64, error) {
	one := id & 0x03
	id >>= 2
	if one != uint64(bits.OnesCount64(id)%4) {
		return 0, ErrBadChecksum
	}

	if g.injectBits > 0 {
		id = extractBits(id, uint(g.injectBits))
	}
	if g.randBits > 0 {
		id = id >> g.randBits
	}

	num := id&0x03 + 1
	mask := 1<<(num*4) - 1
	tabID := (id >> 2) & uint64(mask)
	return tabID, nil
}

// 注入随机数,每n位在低位注入1个
func injectBits(x uint64, bits uint) uint64 {
	bit1 := bits + 1
	mask := uint64(1<<bits - 1)

	r := uint64(rand.Intn(1 << 13))
	t := uint64(0)
	off := uint(0)
	for {
		v := (x&mask)<<1 | (r & 0x01)
		t = v<<off | t
		x >>= bits
		r >>= 1
		off += bit1
		if x == 0 {
			break
		}
	}

	return t
}

// 提取数据,删除随机数
func extractBits(x uint64, bits uint) uint64 {
	bit1 := bits + 1
	mask := uint64(1<<bits - 1)

	t := uint64(0)
	off := uint(0)
	for {
		v := (x >> 1) & mask
		t = v<<off | t
		x >>= bit1
		off += bits
		if x == 0 {
			break
		}
	}

	return t
}

package crypto

import "errors"

var (
	ErrVerifyFail  = errors.New("verify vail,not equal")
	ErrUnknownHash = errors.New("unknown hash")
	ErrNotSupport  = errors.New("not support")
)

type SignType int
type HashType int

const (
	SignRSA SignType = 0 // 数字签名
	SignSum          = 1 // 摘要签名
)

const (
	HashMD5 HashType = iota
	HashSHA1
	HashSha256
)

// 签名算法
//
// 接入第三方平台时,通常需要签名和验签,这里封装了常用的RSA数字签名和摘要签名
// 当需要签名或者验签的数据不是string而且map,url.Values时,可以使用EncodeValues预先处理一下
// 通常的签名过程是:
// 1：去除sign和signType,组成map[string]string
// 2：字典排序,然后进行编码,格式如:bar=baz&foo=quux
// 3: 使用签名算法进行签名,然后标准base64编码
type Signature interface {
	Sign(src string) (string, error)
	Verify(sign string, src string) error
}

type Config struct {
	PublicKey  string
	PrivateKey string
	Secret     string
}

func New(signType SignType, hashType HashType, cfg *Config) (Signature, error) {
	switch signType {
	case SignRSA:
		return NewRSA(hashType, cfg.PublicKey, cfg.PrivateKey)
	case SignSum:
		return NewSum(hashType, cfg.Secret)
	default:
		return nil, ErrNotSupport
	}
}

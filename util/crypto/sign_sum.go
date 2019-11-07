package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

func NewSum(hashType HashType, secret string) (*SummarySignature, error) {
	s := &SummarySignature{}
	if err := s.Init(hashType, secret); err != nil {
		return nil, err
	}

	return s, nil
}

// 摘要签名
type SummarySignature struct {
	hashType HashType
	secret   string
}

func (s *SummarySignature) Init(hashType HashType, secret string) error {
	if len(secret) == 0 {
		return errors.New("invalid secret")
	}
	s.hashType = hashType
	s.secret = secret
	return nil
}

func (s *SummarySignature) Sign(src string) (string, error) {
	var hexSum string
	switch s.hashType {
	case HashMD5:
		// 加入apiKey作加密密钥
		// 这里需要定制么?
		data := fmt.Sprintf("%s&%s", src, s.secret)
		sum := md5.Sum([]byte(data))
		hexSum = hex.EncodeToString(sum[:])
	case HashSHA1:
		h := hmac.New(sha1.New, []byte(s.secret))
		h.Write([]byte(src))
		sum := h.Sum(nil)
		hexSum = hex.EncodeToString(sum[:])
	case HashSha256:
		h := hmac.New(sha256.New, []byte(s.secret))
		h.Write([]byte(src))
		sum := h.Sum(nil)
		hexSum = hex.EncodeToString(sum[:])
	}

	// 有些平台还需要进行大小写转换

	return hexSum, nil
}

func (s *SummarySignature) Verify(sign string, src string) error {
	signRaw, err := s.Sign(src)
	if err != nil {
		return err
	}

	if signRaw != sign {
		return ErrVerifyFail
	}

	return nil
}

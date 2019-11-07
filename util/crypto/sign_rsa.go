package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

func NewRSA(hashType HashType, publicKey, privateKey string) (*RSASignature, error) {
	s := &RSASignature{}
	if err := s.Init(hashType, publicKey, privateKey); err != nil {
		return nil, err
	}

	return s, nil
}

// RSA数字签名算法
type RSASignature struct {
	hash       crypto.Hash
	publicKey  *rsa.PublicKey
	privateKey *rsa.PrivateKey
}

// 初始化数据,私钥和公钥不需要同时存在
// 私钥用于签名,公钥用于验签
func (r *RSASignature) Init(hashType HashType, publicKey, privateKey string) error {
	switch hashType {
	case HashSHA1:
		r.hash = crypto.SHA1
	case HashSha256:
		r.hash = crypto.SHA3_256
	default:
		return ErrUnknownHash
	}

	if publicKey == "" && privateKey == "" {
		return errors.New("invalid config")
	}

	if publicKey != "" {
		if err := r.parsePublicKey(publicKey); err != nil {
			return err
		}
	}

	if privateKey != "" {
		if err := r.parsePrivateKey(privateKey); err != nil {
			return err
		}
	}

	return nil
}

func (r *RSASignature) formatKey(data string, word string) []byte {
	if !strings.HasPrefix(data, "-----BEGIN") {
		data = fmt.Sprintf("-----BEGIN RSASignature %s KEY-----\n%s\n-----END RSASignature %s KEY-----", word, data, word)
	}

	return []byte(data)
}

func (r *RSASignature) parsePublicKey(raw string) error {
	data := r.formatKey(raw, "PUBLIC")
	var block *pem.Block
	block, _ = pem.Decode([]byte(data))
	if block == nil {
		return fmt.Errorf("public key error")
	}

	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err != nil {
		return err
	} else {
		r.publicKey = key.(*rsa.PublicKey)
	}

	return nil
}

func (r *RSASignature) parsePrivateKey(raw string) error {
	data := r.formatKey(raw, "PRIVATE")
	var block *pem.Block
	block, _ = pem.Decode([]byte(data))
	if block == nil {
		return fmt.Errorf("private key error")
	}

	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		r.privateKey = key
		return nil
	}

	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		r.privateKey = key.(*rsa.PrivateKey)
		return nil
	} else {
		return err
	}
}

// 给数据签名,并将结果base64编码
func (r *RSASignature) Sign(src string) (string, error) {
	if r.privateKey == nil {
		return "", errors.New("not init private key")
	}

	h := r.hash.New()
	h.Write([]byte(src))
	s := h.Sum(nil)

	sign, err := rsa.SignPKCS1v15(rand.Reader, r.privateKey, r.hash, s)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sign), nil
}

// 验证签名
func (r *RSASignature) Verify(sign string, data string) error {
	if r.publicKey == nil {
		return errors.New("not init public key")
	}

	decode, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return err
	}

	h := r.hash.New()
	h.Write(decode)
	s := h.Sum(nil)
	return rsa.VerifyPKCS1v15(r.publicKey, r.hash, s, []byte(sign))
}

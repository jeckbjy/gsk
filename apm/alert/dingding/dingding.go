package dingding

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jeckbjy/gsk/apm/alert"
	"github.com/jeckbjy/gsk/util/errorx"
	"github.com/jeckbjy/gsk/util/httpx"
)

func New(opts []*Options) alert.Alert {
	for _, o := range opts {
		if len(o.AtMobiles) > 0 {
			o.atMobileList = strings.Split(o.AtMobiles, ",")
		}
	}

	da := &ddAlert{opts: opts}

	return da
}

type Options struct {
	Tag          string
	URL          string
	Secret       string
	Sign         bool
	AtMobiles    string // 以逗号分隔
	AtAll        bool
	atMobileList []string
}

type ddText struct {
	Content string `json:"content"`
}

type ddAt struct {
	AtMobiles []string `json:"atMobiles"`
	IsAtAll   bool     `json:"isAtAll"`
}

type ddMsgText struct {
	MsgType string `json:"msgtype"`
	Text    ddText `json:"text"`
	At      ddAt   `json:"at"`
}

// https://ding-doc.dingtalk.com/doc#/serverapi2/qf2nxq
type ddAlert struct {
	opts []*Options
}

func (a *ddAlert) Name() string {
	return "dingding"
}

func (a *ddAlert) Send(event *alert.Event) error {
	if event.Type == "" {
		event.Type = alert.TextEvent
	}

	switch event.Type {
	case alert.TextEvent:
		msg := ddMsgText{
			MsgType: "text",
			Text:    ddText{Content: event.Text},
		}
		opt := a.getOptions(event.Tag)
		if opt == nil {
			return errorx.ErrInvalidConfig
		}
		msg.At.AtMobiles = opt.atMobileList
		msg.At.IsAtAll = opt.AtAll
		url := opt.URL
		if opt.Sign && opt.Secret != "" {
			ts := fmt.Sprintf("%+v", time.Now().UnixNano()/int64(time.Millisecond))
			query := map[string]string{
				"timestamp": ts,
				"sign":      doSign(ts, opt.Secret),
			}
			_, err := httpx.Post(url, msg, nil, httpx.QueryMap(query))
			return err
		} else {
			_, err := httpx.Post(url, msg, nil)
			return err
		}
	default:
		return errorx.ErrNotSupport
	}
}

func (a *ddAlert) getOptions(tag string) *Options {
	if tag == "" {
		if len(a.opts) > 0 {
			return a.opts[0]
		}
	} else {
		for _, v := range a.opts {
			if v.Tag == tag {
				return v
			}
		}
	}

	return nil
}

func doSign(ts string, secret string) string {
	src := ts + "\n" + secret
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(src))
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:])
}

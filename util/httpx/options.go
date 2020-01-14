package httpx

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultRetry       = 3
	defaultTimeout     = time.Second * 30
	defaultContentType = TypeJSON
)

type IBackOff interface {
	Reset()
	Next() time.Duration
}

// 用于业务层判断是否需要继续Retry
// 返回true则继续retry,false则不再retry
type RetryFunc func(rsp *http.Response, err error) bool

type Option func(o *Options)
type Options struct {
	Context     context.Context //
	Timeout     time.Duration   // 超时时间
	Retry       int             // 重试次数
	RetryHook   RetryFunc       // 校验是否需要继续Retry
	BackOff     IBackOff        // 每次timeout后等待时间,nil不等待
	ContentType string          // 编码格式
	Charset     string          // 编码格式,utf-8,GBK
	Header      http.Header     // 消息头
	Query       url.Values      // 查询参数
	body        []byte          // Request.Body数据,用于重发
	result      interface{}     // 返回结果
}

func (o *Options) init(opts ...Option) {
	o.Retry = defaultRetry
	o.Timeout = defaultTimeout
	o.ContentType = defaultContentType
	o.Context = context.Background()
	for _, fn := range opts {
		fn(o)
	}
}

func (o *Options) ContentTypeWithCharset() string {
	if len(o.Charset) == 0 {
		return o.ContentType
	}

	return fmt.Sprintf("%s;charset=%s", o.ContentType, o.Charset)
}

// 用于将整个Options,作为参数传递
func (o *Options) Build() Option {
	return func(other *Options) {
		other.Context = o.Context
		other.Timeout = o.Timeout
		other.Retry = o.Retry
		other.BackOff = o.BackOff
		other.ContentType = o.ContentType
		other.Charset = o.Charset
		other.Header = o.Header
		other.Query = o.Query
	}
}

func Timeout(t time.Duration) Option {
	return func(o *Options) {
		o.Timeout = t
	}
}

func Retry(count int, cb RetryFunc) Option {
	return func(o *Options) {
		o.Retry = count
		o.RetryHook = cb
	}
}

func BackOff(b IBackOff) Option {
	return func(o *Options) {
		o.BackOff = b
	}
}

func ContentType(t string) Option {
	return func(o *Options) {
		o.ContentType = t
	}
}

func Charset(x string) Option {
	return func(o *Options) {
		o.Charset = x
	}
}

func Header(header http.Header) Option {
	return func(o *Options) {
		o.Header = header
	}
}

func HeaderMap(header map[string]string) Option {
	return func(o *Options) {
		for k, v := range header {
			o.Header.Add(k, v)
		}
	}
}

func HeaderKV(key, value string) Option {
	return func(o *Options) {
		o.Header.Add(key, value)
	}
}

func Query(q url.Values) Option {
	return func(o *Options) {
		o.Query = q
	}
}

func QueryMap(m map[string]string) Option {
	return func(o *Options) {
		if o.Query == nil {
			o.Query = url.Values{}
		}
		for k, v := range m {
			o.Query.Add(k, v)
		}
	}
}

func QueryKV(key, value string) Option {
	return func(o *Options) {
		o.Query.Add(key, value)
	}
}

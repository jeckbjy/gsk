package httpx

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	ErrNoData      = errors.New("no data")
	ErrNotSupport  = errors.New("not support")
	ErrInvalidType = errors.New("invalid type")
)

const (
	TypeJSON = "application/json"
	TypeXML  = "application/xml"
	TypeForm = "application/x-www-form-urlencoded"
	TypeHTML = "text/html"
	TypeText = "text/plain"
)

const (
	UTF8 = "utf-8"
)

type (
	Request  = http.Request
	Response = http.Response
)

func New(client *http.Client) Client {
	if client == nil {
		// default timeout?
		// https://medium.com/@nate510/don-t-use-go-s-default-http-client-4804cb19f779
		client = &http.Client{Timeout: defaultTimeout}
	}

	return Client{client: client}
}

// 简单封装标准HttpClient
// 支持根据ContentType自动编码和解码
// 支持重试，超时
// 支持添加Header,Query参数
// 其他一些参考:
// https://github.com/hashicorp/go-retryablehttp
// https://github.com/parnurzeal/gorequest
type Client struct {
	client *http.Client
}

// Post 自动编码req和自动解码result,如果为nil,则自动忽略
func (c *Client) Post(url string, req interface{}, result interface{}, opts ...Option) (*Response, error) {
	o := Options{}
	o.init(opts...)
	o.result = result

	body, err := c.encode(o.ContentType, req)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(o.Context, "POST", url, body)
	if err != nil {
		return nil, err
	}

	if len(o.Header) > 0 {
		o.Header.Set("Content-Type", o.ContentType)
	} else {
		request.Header.Set("Content-Type", o.ContentType)
	}
	return c.Do(request, &o)
}

func (c *Client) Get(url string, result interface{}, opts ...Option) (*Response, error) {
	o := Options{}
	o.init(opts...)
	o.result = result
	req, err := http.NewRequestWithContext(o.Context, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req, &o)
}

func (c *Client) Do(req *Request, opts *Options) (*Response, error) {
	if opts.BackOff != nil {
		opts.BackOff.Reset()
	}

	if len(opts.Header) > 0 {
		req.Header = opts.Header
	}

	if len(opts.Query) > 0 {
		query := req.URL.Query()
		for k, v := range query {
			for _, vv := range v {
				opts.Query.Add(k, vv)
			}
		}

		req.URL.RawQuery = opts.Query.Encode()
	}

	for i := 0; ; i++ {
		// reuse request body
		if opts.body != nil {
			req.Body = ioutil.NopCloser(bytes.NewReader(opts.body))
		}

		if opts.Timeout > 0 {
			// 会拷贝一次,能否减少一次拷贝?
			ctx, _ := context.WithTimeout(opts.Context, opts.Timeout)
			req = req.WithContext(ctx)
		}

		rsp, err := c.client.Do(req)
		if err == nil && rsp.StatusCode == http.StatusOK {
			if opts.result != nil {
				err = c.decode(rsp, opts.ContentType, opts.result)
			}

			// 这里已经解析过result了,外部可以使用result结果,用于校验
			if opts.RetryHook == nil || !opts.RetryHook(rsp, err) {
				return rsp, err
			}

			// 需要继续retry
		}

		if i >= opts.Retry {
			return rsp, err
		}

		if opts.BackOff != nil {
			wait := opts.BackOff.Next()
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(wait):
				// timeout continue
			}
		}
	}
}

func (c *Client) encode(contentType string, data interface{}) (io.ReadCloser, error) {
	result, err := Encode(contentType, data)

	if err != nil {
		return nil, err
	}

	if result != nil {
		return ioutil.NopCloser(bytes.NewReader(result)), nil
	}

	return nil, nil
}

func (c *Client) decode(rsp *Response, contentType string, result interface{}) error {
	if val := rsp.Header.Get("Content-Type"); len(val) != 0 {
		ct, _ := ParseContentType(val)
		contentType = ct
	}

	defer rsp.Body.Close()
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	// Reset rsp.Body so it can be use again
	rsp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	// TODO:gbk to utf8 if need
	// http://ju.outofmemory.cn/entry/283132

	return Decode(contentType, body, result)
}

package bi

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
)

func NewHttp() Transport {
	t := &HttpTransport{}
	t.Init()
	return t
}

type HttpTransport struct {
	client *http.Client
}

func (t *HttpTransport) Init() {
	t.client = &http.Client{Timeout: time.Second * 30}
}

func (t *HttpTransport) Send(opts *Options, datas []byte) error {
	rsp, err := t.client.Post(opts.URL, "application/json", bytes.NewReader(datas))
	if err != nil {
		return err
	}

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("send fail,%+v", rsp.Status)
	}

	return nil
}

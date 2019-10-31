package alog

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"sync"
	"time"
)

// 官方的API: https://github.com/elastic/go-elasticsearch
// 但是比较厚重,这里只需要Index
// https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html
// https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html
// TODO:支持Bulk模式,需要ndjson编码
// formatter编码格式必须是json格式
type ElasticChannel struct {
	BaseChannel
	URL      string
	Username string
	Password string
	Timeout  time.Duration // 发送超时,默认10s
	Retry    int           // 失败重试次数,默认不重试,直接丢弃
	Bulk     int           // 用于配置一个批次发送多少条日志,默认1条
	Index    string        // 索引名,按照日期分类?
	queue    Queue         // 单独一个队列
	mux      sync.Mutex    //
	running  bool          // 标识是否已经在发送中
	client   *http.Client  //
}

func (c *ElasticChannel) Name() string {
	return "elastic"
}

func (c *ElasticChannel) SetProperty(key string, value string) error {
	switch key {
	case "url":
		c.URL = value
	case "username":
		c.Username = value
	case "password":
		c.Password = value
	case "timeout":
		if to, err := time.ParseDuration(value); err != nil {
			c.Timeout = to
		} else {
			return err
		}
	case "retry":
		if retry, err := strconv.Atoi(value); err != nil {
			c.Retry = retry
		} else {
			return err
		}
	case "bulk":
		if bulk, err := strconv.Atoi(value); err != nil && bulk > 0 {
			c.Bulk = bulk
		} else {
			return err
		}
	default:
		return c.BaseChannel.SetProperty(key, value)
	}

	return nil
}

func (c *ElasticChannel) Open() error {
	if c.client == nil {
		timeout := c.Timeout
		if c.Timeout == 0 {
			timeout = time.Second * 10
		}
		c.client = &http.Client{
			Timeout: timeout,
		}
	}

	return nil
}

func (c *ElasticChannel) Close() error {
	if c.client != nil {
		c.client.CloseIdleConnections()
		c.client = nil
	}

	return nil
}

// 处理速度可能比较慢,放到单独一个队列中处理
func (c *ElasticChannel) Write(msg *Entry) {
	if c.Open() != nil {
		return
	}

	c.mux.Lock()
	// 丢弃
	if c.queue.Len() >= c.logger.Max() {
		c.mux.Unlock()
		return
	}
	c.queue.Push(msg)
	needRun := false
	if !c.running {
		c.running = true
		needRun = true
	}
	c.mux.Unlock()

	if needRun {
		go c.send()
	}
}

func (c *ElasticChannel) send() {
	for {
		c.mux.Lock()
		if c.queue.Len() == 0 {
			c.running = false
			c.mux.Unlock()
			break
		}
		msg := c.queue.Pop()
		c.mux.Unlock()
		c.doSendIndex(msg)
	}
}

// 根据当前时间按天进行索引
func (c *ElasticChannel) getIndex() string {
	now := time.Now()
	index := fmt.Sprintf("%s_%04d%02d%02d", c.Index, now.Year(), now.Month(), now.Day())
	return index
}

func (c *ElasticChannel) doSendIndex(msg *Entry) {
	// POST /<index>/_doc/
	// POST /<index>/_create/<_id>
	text := msg.Format(c.formatter)
	url := path.Join(c.URL, c.getIndex(), "_doc")
	_ = c.doPost(url, "application/json", text)
}

// POST /_bulk
// POST /<index>/_bulk
// TODO:support ndjson
func (c *ElasticChannel) doSendBulk(msg []*Entry) {
	//url := path.Join(c.URL, c.getIndex(), "_bulk")
	//c.doPost(url, "application/x-ndjson")
}

// 添加重试功能,失败则丢弃
func (c *ElasticChannel) doPost(url, contentType string, data []byte) error {
	var err error
	for i := 0; i < c.Retry+1; i++ {
		_, err = c.client.Post(url, contentType, bytes.NewReader(data))
		if err == nil {
			return nil
		}
	}

	return err
}

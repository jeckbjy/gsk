package db

import (
	"github.com/jeckbjy/gsk/db/driver"
)

func New(driverName string, opts ...OpenOption) (*Client, error) {
	driveri, err := driver.Find(driverName)
	if err != nil {
		return nil, err
	}

	c := &Client{driver: driveri}
	if err := c.Open(opts...); err != nil {
		return nil, err
	}

	return c, nil
}

type Client struct {
	driver driver.Driver
}

func (c *Client) Open(opts ...OpenOption) error {
	o := driver.OpenOptions{}
	for _, fn := range opts {
		fn(&o)
	}

	return c.driver.Open(&o)
}

func (c *Client) Close() error {
	return c.driver.Close()
}

func (c *Client) Ping() error {
	return c.driver.Ping()
}

func (c *Client) Drop(name string) error {
	return c.driver.Drop(name)
}

func (c *Client) Database(name string) (*Database, error) {
	db, err := c.driver.Database(name)
	if err != nil {
		return nil, err
	}

	return &Database{database: db}, nil
}

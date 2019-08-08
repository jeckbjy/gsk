package registry

import (
	"encoding/json"
	"fmt"
	"io"
)

// Service 服务定义,通过Name和Version唯一标识一个服务
// Nodes:注册时只会有一个节点,查询时可能有多个
// Endpoints:为服提供的处理函数,这个是否需要?
// Tags:可以提供一些标签,用于查询等
type Service struct {
	Name      string            `json:"name"`
	Version   string            `json:"version"`
	Meta      map[string]string `json:"meta"`
	Nodes     []*Node           `json:"nodes"`
	Endpoints []*Endpoint       `json:"endpoints"`
}

// FullID 服务唯一标识:服务名加版本号,Service:Version
func (s *Service) FullID() string {
	if s.Version == "" {
		return s.Name
	}

	return fmt.Sprintf("%s:%s", s.Name, s.Version)
}

func (s *Service) First() *Node {
	return s.Nodes[0]
}

// NodeStatus represents the believed status of a member node.
type NodeStatus byte

const (
	// StatusUnknown is the default node status of newly-created nodes.
	StatusUnknown NodeStatus = iota
	// StatusAlive indicates that a node is alive and healthy.
	StatusAlive
	// StatusSuspected indicatates that a node is suspected of being dead.
	StatusSuspected
	// StatusDead indicatates that a node is dead and no longer healthy.
	StatusDead
)

type Node struct {
	Id      string            `json:"id"`      // 唯一ID
	Address string            `json:"address"` // 服务地址[host:port]
	Meta    map[string]string `json:"meta"`    // 附加数据
	Status  NodeStatus        `json:"status"`  // 状态
	conn    io.Closer         `json:"-"`       // 客户端绑定的Conn或者ConnPool,删除节点时会调用Close
}

func (n *Node) IsAlive() bool {
	return n.Status == StatusAlive
}

func (n *Node) Conn() io.Closer {
	return n.conn
}

func (n *Node) SetConn(d io.Closer) {
	n.conn = d
}

type Endpoint struct {
	Name     string            `json:"name"`
	Request  *Value            `json:"request"`
	Response *Value            `json:"response"`
	Meta     map[string]string `json:"meta"`
}

type Value struct {
	Name   string   `json:"name"`
	Type   string   `json:"type"`
	Values []*Value `json:"values"`
}

func Marshal(srv *Service) (string, error) {
	data, err := json.Marshal(srv)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func Unmarshal(data string) (*Service, error) {
	srv := &Service{}
	err := json.Unmarshal([]byte(data), srv)
	if err != nil {
		return nil, err
	}

	return srv, nil
}

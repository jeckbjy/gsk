package registry

import "encoding/json"

func NewService(name string, id string, addr string, meta map[string]string) *Service {
	s := &Service{Name: name, Meta: meta}
	s.Nodes = append(s.Nodes, &Node{id, addr})
	return s
}

// 注册时有且只有1个Node,NodeId则为ServiceId
type Service struct {
	Name      string            `json:"name"`
	Meta      map[string]string `json:"meta"`
	Nodes     []*Node           `json:"nodes"`
	Endpoints []*Endpoint       `json:"endpoints"`
}

func (s *Service) ID() string {
	return s.Nodes[0].Id
}

type Node struct {
	Id      string `json:"id"`
	Address string `json:"address"`
}

type Endpoint struct {
	Name     string `json:"name"`
	Request  *Value `json:"request"`
	Response *Value `json:"response"`
}

type Value struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
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

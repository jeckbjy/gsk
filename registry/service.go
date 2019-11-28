package registry

import "encoding/json"

func NewService(name string, id string, addr string, tags map[string]string) *Service {
	s := &Service{Id: id, Addr: addr, Name: name, Tags: tags}
	return s
}

type Service struct {
	Id        string            `json:"id"`
	Addr      string            `json:"addr"`
	Name      string            `json:"name"`
	Tags      map[string]string `json:"tags"`
	Endpoints []*Endpoint       `json:"endpoints"`
}

func (s *Service) Match(filters map[string]string) bool {
	if len(s.Tags) < len(filters) {
		return false
	}

	for k, v := range filters {
		if s.Tags[k] != v {
			return false
		}
	}

	return true
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

// 默认使用json编码
func (srv *Service) Marshal() (string, error) {
	data, err := json.Marshal(srv)
	return string(data), err
}

func (srv *Service) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), srv)
}

func Unmarshal(data string) (*Service, error) {
	srv := &Service{}
	err := srv.Unmarshal(data)
	return srv, err
}

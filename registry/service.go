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

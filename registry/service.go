package registry

import "encoding/json"

func NewService(name string, id string, addr string, tags map[string]string) *Service {
	s := &Service{Id: id, Addr: addr, Name: name, Tags: tags}
	return s
}

// TODO:添加其他信息,比如Version,Zone,Endpoint等信息
type Service struct {
	Id   string            `json:"id"`
	Addr string            `json:"addr"`
	Name string            `json:"name"`
	Tags map[string]string `json:"tags"`
}

/// 检测Service是否完全满足filter条件
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

// 默认使用json编码
func (s *Service) Marshal() string {
	data, _ := json.Marshal(s)
	return string(data)
}

func (s *Service) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), s)
}

func Unmarshal(data string) (*Service, error) {
	srv := &Service{}
	err := srv.Unmarshal(data)
	return srv, err
}

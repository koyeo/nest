package protocol

import (
	"fmt"
)

type ServerManager struct {
	_list []*Server
	_map  map[string]*Server
}

func (p *ServerManager) Add(item *Server) error {
	
	if item.Name == "" {
		return fmt.Errorf("server with empty name")
	}
	
	if p._map == nil {
		p._map = map[string]*Server{}
	}
	
	if _, ok := p._map[item.Name]; ok {
		return fmt.Errorf("duplicated server: %s", item.Name)
	}
	
	p._map[item.Name] = item
	p._list = append(p._list, item)
	
	return nil
}

func (p *ServerManager) Get(name string) *Server {
	if p._map == nil {
		return nil
	}
	return p._map[name]
}

func (p *ServerManager) List() []*Server {
	return p._list
}

type Server struct {
	Name         string
	Host         string
	Port         uint64
	User         string
	Password     string
	IdentityFile string
}

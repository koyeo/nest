package core

import (
	"fmt"
)

type WatcherManager struct {
	_list []*Watcher
	_map  map[string]*Watcher
}

func (p *WatcherManager) Add(item *Watcher) error {
	
	if item.Name == "" {
		return fmt.Errorf("watcher with empty name")
	}
	
	if p._map == nil {
		p._map = map[string]*Watcher{}
	}
	
	if _, ok := p._map[item.Name]; ok {
		return fmt.Errorf("duplicated watcher: %s", item.Name)
	}
	
	p._map[item.Name] = item
	p._list = append(p._list, item)
	
	return nil
}

func (p *WatcherManager) Get(name string) *Watcher {
	if p._map == nil {
		return nil
	}
	return p._map[name]
}

func (p *WatcherManager) List() []*Watcher {
	return p._list
}

type Watcher struct {
	Command string
	Watch   []string
	Name    string
}

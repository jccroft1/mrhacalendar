package main

import "sync"

var cache = struct {
	sync.RWMutex
	data map[string]*Calendar
}{
	data: make(map[string]*Calendar),
}

func get(team string) *Calendar {
	cache.RLock()
	defer cache.RUnlock()
	return cache.data[team]
}

func set(team string, c *Calendar) {
	cache.Lock()
	defer cache.Unlock()
	cache.data[team] = c
}

package main

import (
	"sync"
	"time"
)

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
	// expire cache after 12 hours
	// setting again will reset cache
	go func() {
		time.Sleep(12 * time.Hour)
		remove(team)
	}()
}

func remove(team string) {
	cache.Lock()
	defer cache.Unlock()
	delete(cache.data, team)
}

// Provides a concurrency-safe cache of a function of
// a function.  Requests for different keys proceed in parallel.
// Concurrent requests for the same key block until the first completes.
// This implementation uses a Mutex.
package main

import (
	"log"
	"sync"
	"time"
)

var cache *Cache

func init() {
	cache = &Cache{f: extract, data: make(map[string]*entry)}
}

// Func is the type of the function to memoize.
type Func func(string) (*Calendar, error)

type result struct {
	value interface{}
	err   error
}

//!+
type entry struct {
	res   result
	ready chan struct{} // closed when res is ready
}

// Cache struct
type Cache struct {
	f          Func
	sync.Mutex // guards cache
	data       map[string]*entry
}

// Get used to retreive cached entries
func (cache *Cache) Get(key string) (value interface{}, err error) {
	cache.Lock()
	e := cache.data[key]

	if e == nil {
		// This is the first request for this key.
		// This goroutine becomes responsible for computing
		// the value and broadcasting the ready condition.
		e = &entry{ready: make(chan struct{})}
		cache.data[key] = e

		e.res.value, e.res.err = cache.f(key)
		cache.Unlock()

		close(e.ready) // broadcast ready condition

		// invalidate cache after 12 hours
		go func() {
			time.Sleep(12 * time.Hour)
			cache.Lock()
			defer cache.Unlock()
			delete(cache.data, key)
		}()
	} else {
		log.Println("found cached entry")
		// This is a repeat request for this key.
		cache.Unlock()

		<-e.ready // wait for ready condition
	}
	return e.res.value, e.res.err
}

//!-

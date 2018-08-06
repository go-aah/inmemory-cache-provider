// Copyright (c) Jeevanandam M. (https://github.com/jeevatkm)
// aahframework.org/cache/inmemory source code and usage is governed by a MIT style
// license that can be found in the LICENSE file.

package inmemory // import "aahframework.org/cache/inmemory"

import (
	"sync"
	"time"

	"aahframework.org/aah.v0/cache"
	"aahframework.org/config.v0"
	"aahframework.org/log.v0"
)

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Provider type and implements cache.Provider interface
//______________________________________________________________________________

// Provider struct represents the in-memory cache provider.
type Provider struct {
	name string
	cfg  *cache.Config
}

var _ cache.Provider = (*Provider)(nil)

// Init method is not applicable for in-memory cache provider.
func (p *Provider) Init(name string, _ *config.Config, _ log.Loggerer) error {
	p.name = name
	// nothing to initialize for in-memory cache
	return nil
}

// Create method creates new in-memory cache with given options.
func (p *Provider) Create(cfg *cache.Config) (cache.Cache, error) {
	p.cfg = cfg
	c := &InMemory{
		cfg:     p.cfg,
		mu:      sync.RWMutex{},
		entries: make(map[string]entry),
	}

	if p.cfg.EvictionMode != cache.EvictionModeNoTTL {
		go c.startSweeper()
	}

	return c, nil
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// InMemory cache and implements cache.Cache interface
//______________________________________________________________________________

// InMemory struct represents in-memory cache implementation.
type InMemory struct {
	cfg     *cache.Config
	mu      sync.RWMutex
	entries map[string]entry
}

var _ cache.Cache = (*InMemory)(nil)

// Name method returns the cache store name.
func (im *InMemory) Name() string {
	return im.cfg.Name
}

// Get method returns the cached entry for given key if it exists otherwise nil.
func (im *InMemory) Get(k string) interface{} {
	im.mu.RLock()
	e, f := im.entries[k]
	if f && !e.IsExpired() {
		im.mu.RUnlock()
		if im.cfg.EvictionMode == cache.EvictionModeSlide {
			if e.D > 0 {
				e.E = time.Now().Add(e.D).UnixNano()
				im.mu.Lock()
				im.entries[k] = e
				im.mu.Unlock()
			}
		}
		return e.V
	}
	im.mu.RUnlock()
	return nil
}

// GetOrPut method returns the cached entry for the given key if it exists otherwise
// it puts the new entry into cache store and returns the value.
func (im *InMemory) GetOrPut(k string, v interface{}, d time.Duration) interface{} {
	ev := im.Get(k)
	if ev == nil {
		_ = im.put(k, v, d)
		return v
	}
	return ev
}

// Put method adds the cache entry with specified expiration. Returns error
// if cache entry exists.
func (im *InMemory) Put(k string, v interface{}, d time.Duration) error {
	return im.put(k, v, d)
}

// Delete method deletes the cache entry from cache store.
func (im *InMemory) Delete(k string) {
	im.mu.Lock()
	delete(im.entries, k)
	im.mu.Unlock()
}

// Exists method checks given key exists in cache store and its not expried.
func (im *InMemory) Exists(k string) bool {
	return im.Get(k) != nil
}

// Flush methods flushes(deletes) all the cache entries from cache.
func (im *InMemory) Flush() {
	im.mu.Lock()
	im.entries = make(map[string]entry)
	im.mu.Unlock()
}

//‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾‾
// Cache type's unexported methods
//______________________________________________________________________________

func (im *InMemory) put(k string, v interface{}, d time.Duration) error {
	if ev := im.Get(k); ev == nil {
		var exp int64
		if d > 0 {
			exp = time.Now().Add(d).UnixNano()
		}
		im.mu.Lock()
		im.entries[k] = entry{V: v, D: d, E: exp}
		im.mu.Unlock()
		return nil
	}
	return cache.ErrEntryExists
}

func (im *InMemory) startSweeper() {
	ticker := time.NewTicker(im.cfg.SweepInterval)
	for {
		<-ticker.C
		nt := time.Now().UnixNano()
		im.mu.Lock()
		for k, e := range im.entries {
			if e.E > 0 && nt > e.E {
				delete(im.entries, k)
			}
		}
		im.mu.Unlock()
	}
}

type entry struct {
	E int64
	D time.Duration
	V interface{}
}

func (i entry) IsExpired() bool {
	return i.E > 0 && time.Now().UnixNano() > i.E
}

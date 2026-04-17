package engine

import "sync"

// Cache is a concurrency-safe store for compiled *Schema values keyed by
// an arbitrary string (typically a JSON pointer into the source
// document, or the fully-qualified component name). Use it to avoid
// recompiling the same schema once per instance validation — compilation
// is the expensive part of the hot path.
//
// Cache entries are stored by value (pointer) so reuse across goroutines
// is safe. The underlying santhosh *jsonschema.Schema is read-only at
// validate time.
type Cache struct {
	mu    sync.RWMutex
	items map[string]*Schema
}

// NewCache returns an empty cache.
func NewCache() *Cache {
	return &Cache{items: make(map[string]*Schema)}
}

// Get looks up a compiled schema by key.
func (c *Cache) Get(key string) (*Schema, bool) {
	if c == nil {
		return nil, false
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	s, ok := c.items[key]
	return s, ok
}

// Put stores a compiled schema under the given key. Subsequent Get calls
// for the same key will return the cached value until Invalidate or
// Reset is called.
func (c *Cache) Put(key string, s *Schema) {
	if c == nil || s == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = s
}

// GetOrCompile returns the cached schema for key if present, otherwise
// compiles `raw` using the supplied options, caches the result, and
// returns it.
func (c *Cache) GetOrCompile(key string, raw any, opts CompileOpts) (*Schema, error) {
	if s, ok := c.Get(key); ok {
		return s, nil
	}
	s, err := Compile(raw, opts)
	if err != nil {
		return nil, err
	}
	c.Put(key, s)
	return s, nil
}

// Invalidate removes a single entry.
func (c *Cache) Invalidate(key string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Reset empties the cache.
func (c *Cache) Reset() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]*Schema)
}

// Len returns the number of cached entries (useful for tests and
// diagnostics).
func (c *Cache) Len() int {
	if c == nil {
		return 0
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

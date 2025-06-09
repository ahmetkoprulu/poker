package api

import (
	"sync"
)

type Factory struct {
	clients map[string]*ApiClient
	mu      sync.RWMutex
}

func NewFactory() *Factory {
	return &Factory{
		clients: make(map[string]*ApiClient),
	}
}

func (f *Factory) GetOrCreate(name string, config ClientConfig) *ApiClient {
	f.mu.RLock()
	client, exists := f.clients[name]
	f.mu.RUnlock()

	if exists {
		return client
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Double check after acquiring write lock
	if client, exists = f.clients[name]; exists {
		return client
	}

	client = NewApiClient(config)
	f.clients[name] = client
	return client
}

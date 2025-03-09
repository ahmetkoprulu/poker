package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

type ApiClient struct {
	baseUrl string
	client  *http.Client
	headers map[string]string
	mu      sync.RWMutex
}

func NewApiClient(config ClientConfig) *ApiClient {
	dialer := &net.Dialer{
		Timeout:   config.Timeout,
		KeepAlive: config.KeepAlive,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          config.MaxIdleConns,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		ForceAttemptHTTP2:     true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout,
	}

	return &ApiClient{
		client:  client,
		baseUrl: config.BaseURL,
		headers: config.Headers,
	}
}

func (c *ApiClient) SetHeader(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers[key] = value
}

func (c *ApiClient) RemoveHeader(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.headers, key)
}

func (c *ApiClient) Get(ctx context.Context, path string, result interface{}) error {
	return c.Request(ctx, http.MethodGet, path, nil, result)
}

func (c *ApiClient) Post(ctx context.Context, path string, payload, result interface{}) error {
	return c.Request(ctx, http.MethodPost, path, payload, result)
}

func (c *ApiClient) Put(ctx context.Context, path string, payload, result interface{}) error {
	return c.Request(ctx, http.MethodPut, path, payload, result)
}

func (c *ApiClient) Patch(ctx context.Context, path string, payload, result interface{}) error {
	return c.Request(ctx, http.MethodPatch, path, payload, result)
}

func (c *ApiClient) Delete(ctx context.Context, path string, result interface{}) error {
	return c.Request(ctx, http.MethodDelete, path, nil, result)
}

func (c *ApiClient) Request(ctx context.Context, method, path string, payload, result interface{}) error {
	// url := c.baseUrl + path
	url := path
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		body = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	c.mu.RLock()
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}
	c.mu.RUnlock()

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	if result != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}

func (c *ApiClient) Close() {
	type closeIdler interface {
		CloseIdleConnections()
	}

	if tr, ok := c.client.Transport.(closeIdler); ok {
		tr.CloseIdleConnections()
	}
}

type ApiResponse[T any] struct {
	Success bool   `json:"success"`
	Status  int    `json:"status"`
	Data    T      `json:"data"`
	Message string `json:"message"`
}

func DefaultConfig() ClientConfig {
	return ClientConfig{
		Timeout:               30 * time.Second,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		MaxConnsPerHost:       100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		KeepAlive:             30 * time.Second,
		Headers:               make(map[string]string),
	}
}

type ClientConfig struct {
	BaseURL               string
	Timeout               time.Duration
	MaxIdleConns          int
	MaxIdleConnsPerHost   int
	MaxConnsPerHost       int
	IdleConnTimeout       time.Duration
	TLSHandshakeTimeout   time.Duration
	ExpectContinueTimeout time.Duration
	KeepAlive             time.Duration
	Headers               map[string]string
}

package transport

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httputil"
	"sync"
)

type Cache interface {
	Get(key string) (responseBytes []byte, ok bool)
	Set(key string, responseBytes []byte)
	Delete(key string)
}

func cacheKey(req *http.Request) string {
	return req.URL.String()
}

type MemoryCache struct {
	mu    sync.RWMutex
	items map[string][]byte
}

func (c *MemoryCache) Get(key string) (resp []byte, ok bool) {
	c.mu.RLock()
	resp, ok = c.items[key]
	c.mu.RUnlock()
	return resp, ok
}

func (c *MemoryCache) Set(key string, resp []byte) {
	c.mu.Lock()
	c.items[key] = resp
	c.mu.Unlock()
}

func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func NewMemoryCache() *MemoryCache {
	c := &MemoryCache{items: map[string][]byte{}}
	return c
}

type Transport struct {
	Token     string
	Transport http.RoundTripper
	Cache     Cache
	mu        sync.RWMutex
	modReq    map[*http.Request]*http.Request
}

func NewTransport(c Cache) *Transport {
	return &Transport{
		Cache: c,
	}
}

func (t *Transport) Client() *http.Client {
	return &http.Client{Transport: t}
}

func CachedResponse(c Cache, req *http.Request) (resp *http.Response, err error) {
	cachedVal, ok := c.Get(cacheKey(req))
	if !ok {
		return
	}

	b := bytes.NewBuffer(cachedVal)
	return http.ReadResponse(bufio.NewReader(b), req)
}

func (t *Transport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	cacheKey := cacheKey(req)
	cacheable := req.Method == "GET" || req.Method == "HEAD"

	var cachedResp *http.Response
	if cacheable {
		cachedResp, err = CachedResponse(t.Cache, req)
	} else {
		t.Cache.Delete(cacheKey)
	}

	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	if t.Token != "" {
		req.Header.Set("Authorization", "Bearer "+t.Token)
	}

	if cacheable && cachedResp != nil {
		etag := cachedResp.Header.Get("etag")
		if etag != "" && req.Header.Get("etag") == "" {
			req.Header.Set("if-none-match", etag)
		}

		resp, err = transport.RoundTrip(req)
		if err == nil && req.Method == "GET" && resp.StatusCode == http.StatusNotModified {
			cachedResp.Status = resp.Status
			return cachedResp, nil
		} else {
			return resp, err
		}
	} else {
		resp, err = transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}
	}

	if cacheable {
		respBytes, err := httputil.DumpResponse(resp, true)
		if err == nil {
			t.Cache.Set(cacheKey, respBytes)
		}
	}

	return resp, nil
}

func (t *Transport) SetToken(token string) *Transport {
	t.Token = token
	return t
}

func NewMemoryCacheTransport() *Transport {
	c := NewMemoryCache()
	t := NewTransport(c)
	return t
}

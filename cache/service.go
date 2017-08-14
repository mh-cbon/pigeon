package cache

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"

	"google.golang.org/api/googleapi"
	vision "google.golang.org/api/vision/v1"
)

// Call with cache support
func Call(s *vision.ImagesAnnotateCall, store Storer, batch *vision.BatchAnnotateImagesRequest) *ServiceCall {
	return &ServiceCall{ImagesAnnotateCall: s, store: store, req: batch}
}

// ServiceCall wraps a vison.ServiceCall with a cache.
type ServiceCall struct {
	*vision.ImagesAnnotateCall
	store       Storer
	req         *vision.BatchAnnotateImagesRequest
	maxPerMonth int
}

func (c *ServiceCall) getHash() (string, error) {
	res, err := json.Marshal(c.req)
	if err != nil {
		return "", err
	}

	h := sha1.New()
	h.Write(res)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

// MaxPerMonth queries
func (c *ServiceCall) MaxPerMonth(n int) *ServiceCall {
	c.maxPerMonth = n
	return c
}

// Do reads from cache or execute online.
func (c *ServiceCall) Do(opts ...googleapi.CallOption) (*vision.BatchAnnotateImagesResponse, error) {
	hash, err := c.getHash()
	if err != nil {
		return nil, err
	}

	if res, ok := c.store.Get(hash); ok {
		return res, nil
	}

	if c.maxPerMonth > 0 {
		n, err2 := c.store.LastMonthCount()
		if err2 != nil {
			return nil, err
		}
		if n+1 > c.maxPerMonth {
			return nil, fmt.Errorf("max limit of %v monthly queries exceeded", c.maxPerMonth)
		}
	}
	res, err := c.ImagesAnnotateCall.Do(opts...)
	if err != nil {
		return nil, err
	}
	c.store.Save(hash, res)
	return res, nil
}

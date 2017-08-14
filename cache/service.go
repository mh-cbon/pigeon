package cache

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"

	"google.golang.org/api/googleapi"
	vision "google.golang.org/api/vision/v1"
)

// ImagesService can cache
type ImagesService struct {
	store Storer
	*vision.ImagesService
}

// Annotate returns from cache or remote.
func (v ImagesService) Annotate(batchannotateimagesrequest *vision.BatchAnnotateImagesRequest) AnnotationDoer {
	return &ServiceCall{
		store:             v.store,
		visionServiceCall: &visionServiceCall{v.ImagesService.Annotate(batchannotateimagesrequest)},
		req:               batchannotateimagesrequest,
	}
}

// ServiceCall wraps a vison.ServiceCall with a cache.
type ServiceCall struct {
	*visionServiceCall
	store Storer
	req   *vision.BatchAnnotateImagesRequest
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

// Do reads from cache or execute online.
func (c *ServiceCall) Do(opts ...googleapi.CallOption) (*vision.BatchAnnotateImagesResponse, error) {
	hash, err := c.getHash()
	if err != nil {
		return nil, err
	}

	if res, ok := c.store.Get(hash); ok {
		return res, nil
	}
	res, err := c.visionServiceCall.Do(opts...)
	if err != nil {
		return nil, err
	}
	c.store.Save(hash, res)
	return res, nil
}

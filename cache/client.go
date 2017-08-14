package cache

import (
	"net/http"

	context "golang.org/x/net/context"

	"github.com/kaneshin/pigeon"
	"google.golang.org/api/googleapi"
	vision "google.golang.org/api/vision/v1"
)

// Client with Annotater wrapper.
type Client struct {
	*pigeon.Client
	annotater Annotater
}

// New returns a pointer to a new Client object.
func New(s Storer, c *pigeon.Config, httpClient ...*http.Client) (*Client, error) {
	p, err := pigeon.New(c, httpClient...)
	if err != nil {
		return nil, err
	}
	return &Client{
		Client: p,
		annotater: ImagesService{
			store:         s,
			ImagesService: p.ImagesService(),
		},
	}, nil
}

// ImagesService returns a pointer to an Annotater.
func (c Client) ImagesService() Annotater {
	return c.annotater
}

// Annotater is an *vision.ImagesService
type Annotater interface {
	Annotate(batchannotateimagesrequest *vision.BatchAnnotateImagesRequest) AnnotationDoer
}

type visionServiceAnnotate struct {
	*vision.ImagesService
}

// Annotate returns an AnnotationDoer rather than *vision.ImagesAnnotateCall.
func (v visionServiceAnnotate) Annotate(batchannotateimagesrequest *vision.BatchAnnotateImagesRequest) AnnotationDoer {
	return &visionServiceCall{v.ImagesService.Annotate(batchannotateimagesrequest)}
}

type visionServiceCall struct {
	*vision.ImagesAnnotateCall
}

// Context sets the context to be used in this call's Do method. Any
// pending HTTP request will be aborted if the provided context is
// canceled.
func (c *visionServiceCall) Context(ctx context.Context) AnnotationDoer {
	c.ImagesAnnotateCall.Context(ctx)
	return c
}

// AnnotationDoer is an *vision.ImagesAnnotateCall
type AnnotationDoer interface {
	Context(ctx context.Context) AnnotationDoer
	Do(opts ...googleapi.CallOption) (*vision.BatchAnnotateImagesResponse, error)
	// Fields(s ...googleapi.Field) *vision.ImagesAnnotateCall
	// Header() http.Header
}

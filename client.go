package restful

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// NewObjectFunc is a function type which returns a new object for deserialization into.
type NewObjectFunc func() interface{}

// Client represents a RESTful HTTP client.
type Client struct {
	ClientCodec
	HTTPClient
	URL         string
	NewFunc     NewObjectFunc
	NewListFunc NewObjectFunc
}

// Get retrieves the record with id id from an endpoint.
func (c Client) Get(id ID) (interface{}, error) {
	resp, err := c.do("GET", c.getEndpoint(id), nil)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Body == nil {
		return nil, nil // FIXME no response body from server?
	}
	item := c.New()
	err = c.Decode(resp.Body, item)
	return item, err
}

// GetAll retrieves all record from an endpoint.
func (c Client) GetAll() (interface{}, error) {
	resp, err := c.do("GET", c.URL, nil)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Body == nil {
		return nil, nil // FIXME no response body from server?
	}
	list := c.NewList()
	// FIXME need to check if resp.Body is nil?
	err = c.Decode(resp.Body, list)
	return list, err
}

// Put creates a record at an endpoint and returns the resource id.
func (c Client) Put(v interface{}) (interface{}, error) {
	var b bytes.Buffer
	err := c.Encode(&b, v)
	if err != nil {
		return "", err
	}
	resp, err := c.do("POST", c.URL, &b)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Body == nil {
		return nil, nil // FIXME no response body from server?
	}
	item := c.New()
	err = c.Decode(resp.Body, item)
	return item, err
}

// PutWithID updates a record with a given id at an endpoint.
func (c Client) Update(id ID, v interface{}) error {
	var b bytes.Buffer
	err := c.Encode(&b, v)
	if err != nil {
		return err
	}
	_, err = c.do("POST", c.getEndpoint(id), &b)
	return err
}

// Delete deletes a record with a given id at an endpoint.
func (c Client) Delete(id ID) error {
	_, err := c.do("DELETE", c.getEndpoint(id), nil)
	return err
}

// New allocates and returns an empty record for deserialization.
func (c Client) New() interface{} {
	return c.NewFunc()
}

// NewList allocates and returns an empty list of records for deserialization.
func (c Client) NewList() interface{} {
	return c.NewListFunc()
}

func (c Client) do(method, urlStr string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", c.GetBodyType())
	}
	return c.Do(req)
}

func (c Client) getEndpoint(id ID) string {
	return strings.Join([]string{c.URL, string(id)}, "/")
}

// NewJSONClient creates a new RESTful JSON client.
func NewJSONClient(url string, newFunc, newListFunc NewObjectFunc) *Client {
	return &Client{
		ClientCodec: JSONCodec,
		HTTPClient:  http.DefaultClient,
		NewFunc:     newFunc,
		NewListFunc: newListFunc,
	}
}

var _ ClientStore = &Client{}

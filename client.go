package restful

import (
	"bytes"
	"net/http"
	"strings"
)

// NewObjectFunc is a function type which returns a new object for deserialization into.
type NewObjectFunc func() interface{}

// Client represents a RESTful HTTP client.
type Client struct {
	ClientCodec
	URL         string
	NewFunc     NewObjectFunc
	NewListFunc NewObjectFunc
}

// Get retrieves the record with id id from an endpoint.
func (c Client) Get(id ID) (interface{}, error) {
	resp, err := http.Get(c.getEndpoint(id))
	if err != nil {
		return nil, err
	}
	item := c.New()
	// FIXME need to check if resp.Body is nil?
	err = c.Decode(resp.Body, item)
	return item, err
}

// GetAll retrieves all record from an endpoint.
func (c Client) GetAll() (interface{}, error) {
	resp, err := http.Get(c.URL)
	if err != nil {
		return nil, err
	}
	list := c.NewList()
	// FIXME need to check if resp.Body is nil?
	err = c.Decode(resp.Body, list)
	return list, err
}

// Put creates a record at an endpoint and returns the resource id.
func (c Client) Put(v interface{}) (ID, error) {
	var b bytes.Buffer
	err := c.Encode(&b, v)
	if err != nil {
		return "", err
	}
	resp, err := http.Post(c.URL, c.GetBodyType(), &b)
	// FIXME need to check if resp.Body is nil?
	var id ID // FIXME is this the right response
	err = c.Decode(resp.Body, &id)
	return id, err
}

// PutWithID updates a record with a given id at an endpoint.
func (c Client) Update(id ID, v interface{}) error {
	var b bytes.Buffer
	err := c.Encode(&b, v)
	if err != nil {
		return err
	}
	_, err = http.Post(c.getEndpoint(id), c.GetBodyType(), &b)
	return err
}

// Delete deletes a record with a given id at an endpoint.
func (c Client) Delete(id ID) error {
	req, err := http.NewRequest("DELETE", c.getEndpoint(id), nil)
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
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

func (c Client) getEndpoint(id ID) string {
	return strings.Join([]string{c.URL, string(id)}, "/")
}

// NewJSONClient creates a new RESTful JSON client.
func NewJSONClient(url string, newFunc, newListFunc NewObjectFunc) *Client {
	return &Client{
		ClientCodec: JSONCodec,
		NewFunc:     newFunc,
		NewListFunc: newListFunc,
	}
}

var _ ClientStore = &Client{}

package restful

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// ErrNoResponseBody is the error returned when there is no response body.
var ErrNoResponseBody = errors.New("no response body from server")

// ErrStatus represents a status code error.
// It is returned when a client method gets a status code other than 404 or 200.
type ErrStatus struct {
	Status     string
	StatusCode int
}

func (e ErrStatus) Error() string {
	return fmt.Sprintf("%v %v", e.StatusCode, e.Status)
}

// HTTPClient is an interface that wraps the Do method.
// DefaultClient in the "net/http" implements the HTTPClient interface.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// NewObjectFunc is a function type which returns a new object for
// deserialization into.
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
	if resp.Body == nil {
		return nil, ErrNoResponseBody
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrMissing{id}
	} else if resp.StatusCode != http.StatusOK {
		return nil, ErrStatus{Status: resp.Status, StatusCode: resp.StatusCode}
	}
	item := c.New()
	err = c.Decode(resp.Body, item)
	return item, err
}

// GetAll retrieves all record from an endpoint.
func (c Client) GetAll() (interface{}, error) {
	log.Println("GetAll()")
	resp, err := c.do("GET", c.URL, nil)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, ErrStatus{Status: resp.Status, StatusCode: resp.StatusCode}
	}
	if resp.Body == nil {
		return nil, ErrNoResponseBody // FIXME return empty list?
	}
	list := c.NewList()
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
	if resp.StatusCode != http.StatusOK {
		return nil, ErrStatus{Status: resp.Status, StatusCode: resp.StatusCode}
	}
	if resp.Body == nil {
		return nil, ErrNoResponseBody
	}
	item := c.New()
	err = c.Decode(resp.Body, item)
	return item, err
}

// Update updates a record with a given id at an endpoint.
func (c Client) Update(id ID, v interface{}) error {
	var b bytes.Buffer
	err := c.Encode(&b, v)
	if err != nil {
		return err
	}
	resp, err := c.do("POST", c.getEndpoint(id), &b)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrMissing{id}
	} else if resp.StatusCode != http.StatusOK {
		return ErrStatus{Status: resp.Status, StatusCode: resp.StatusCode}
	}
	return nil
}

// Delete deletes a record with a given id at an endpoint.
func (c Client) Delete(id ID) error {
	resp, err := c.do("DELETE", c.getEndpoint(id), nil)
	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		return ErrMissing{id}
	} else if resp.StatusCode != http.StatusOK {
		return ErrStatus{Status: resp.Status, StatusCode: resp.StatusCode}
	}
	return nil
}

// New allocates and returns an empty record for deserialization.
func (c Client) New() interface{} {
	return c.NewFunc()
}

// NewList allocates and returns an empty list of records for deserialization.
func (c Client) NewList() interface{} {
	return c.NewListFunc()
}

func (c Client) do(
	method, urlStr string,
	body io.Reader,
) (*http.Response, error) {
	log.Printf("do(%v, %v, %v)", method, urlStr, body)
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
		URL:         url,
		NewFunc:     newFunc,
		NewListFunc: newListFunc,
	}
}

var _ ClientStore = &Client{}

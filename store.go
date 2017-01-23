// Package restful provides a RESTful HTTP client and server implementation.
package restful

import "fmt"

// ErrMissing represents a missing ID error for get/update/delete operations
// on a Store.
type ErrMissing struct {
	ID
}

func (e ErrMissing) Error() string {
	return fmt.Sprintf("not found: %q", e.ID)
}

// ID represents a record ID in a Store
type ID string

// Store presents an interface to a data store.
// NOTE GetAll() should return a non-nil value even when there are no records
// returned.
type Store interface {
	Put(interface{}) (interface{}, error)
	Update(ID, interface{}) error
	Get(ID) (interface{}, error)
	GetAll() (interface{}, error)
	Delete(ID) error
	New() interface{}
}

// ClientStore extends Store with a NewList method.
type ClientStore interface {
	Store
	NewList() interface{}
}

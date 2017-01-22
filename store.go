package restful

// ID represents a record ID in a Store
type ID string

// Store is presents an interface to a data store.
type Store interface {
	Put(interface{}) (ID, error)
	PutWithId(ID, interface{}) error
	Delete(ID) error
	Get(ID) (interface{}, error)
	GetAll() (interface{}, error)
	New() interface{}
}

// ClientStore extends Store with a NewList method.
type ClientStore interface {
	Store
	NewList() interface{}
}

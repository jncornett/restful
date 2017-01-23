package restful

import "sync"

// RWStore wraps another Store, providing concurrency-safe
// Read/write access to the underlying store.
type RWStore struct {
	store Store
	sync.RWMutex
}

func NewRWStore(s Store) *RWStore {
	return &RWStore{store: s}
}

func (s *RWStore) Delete(id ID) error {
	s.Lock()
	defer s.Unlock()
	return s.store.Delete(id)
}

func (s *RWStore) Get(id ID) (interface{}, error) {
	s.RLock()
	defer s.RUnlock()
	return s.store.Get(id)
}

func (s *RWStore) GetAll() (interface{}, error) {
	s.RLock()
	defer s.RUnlock()
	return s.store.GetAll()
}

func (s *RWStore) New() interface{} {
	return s.store.New()
}

func (s *RWStore) Put(v interface{}) (interface{}, error) {
	s.Lock()
	defer s.Unlock()
	return s.store.Put(v)
}

func (s *RWStore) Update(id ID, v interface{}) error {
	s.Lock()
	defer s.Unlock()
	return s.store.Update(id, v)
}

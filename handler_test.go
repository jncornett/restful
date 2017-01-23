package restful_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jncornett/restful"
)

type dummyObject struct {
	ID      restful.ID `json:"id"`
	Payload string     `json:"payload"`
}

type dummyStore map[restful.ID]dummyObject

func (s dummyStore) Get(id restful.ID) (interface{}, error) {
	item, ok := s[id]
	if ok {
		return &item, nil
	}
	return nil, restful.ErrMissing{id}
}

func (s dummyStore) GetAll() (interface{}, error) {
	var out []dummyObject
	for _, obj := range s {
		out = append(out, obj)
	}
	return out, nil
}

func (s dummyStore) Put(v interface{}) (interface{}, error) {
	obj, ok := v.(*dummyObject)
	if !ok {
		return nil, errors.New("Wrong type")
	}
	s[obj.ID] = *obj
	return &obj, nil
}

func (s dummyStore) Update(id restful.ID, v interface{}) error {
	_, ok := s[id]
	if !ok {
		return restful.ErrMissing{id}
	}
	obj, ok := v.(*dummyObject)
	if !ok {
		return errors.New("wrong type")
	}
	s[id] = *obj
	return nil
}

func (s dummyStore) Delete(id restful.ID) error {
	_, ok := s[id]
	if !ok {
		return restful.ErrMissing{id}
	}
	delete(s, id)
	return nil
}

func (s dummyStore) New() interface{} {
	return &dummyObject{}
}

func newTestServer(s restful.Store) *httptest.Server {
	return httptest.NewServer(restful.NewJSONHandler(s))
}

func newTestClient(url string) *restful.Client {
	return restful.NewJSONClient(
		url,
		func() interface{} { return &dummyObject{} },
		func() interface{} { return []dummyObject{} },
	)
}

func TestHandler_Get(t *testing.T) {
	var (
		testID  = restful.ID("a")
		testObj = dummyObject{ID: testID, Payload: "hello"}
		store   = dummyStore{testID: testObj}
	)
	server := newTestServer(store)
	defer server.Close()
	resp, err := http.Get(server.URL + "/a")
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusOK != resp.StatusCode {
		t.Errorf("expected status to be %v, got %v", http.StatusNotFound, resp.StatusCode)
	}
	if resp.Body == nil {
		t.Fatal("expected non nil response body")
	}
	var obj dummyObject
	err = json.NewDecoder(resp.Body).Decode(&obj)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(testObj, obj) {
		t.Errorf("expected obj to be %v, got %v", testObj, obj)
	}
}

func TestHandler_GetMissing(t *testing.T) {
	var store = &dummyStore{}
	server := newTestServer(store)
	defer server.Close()
	resp, err := http.Get(server.URL + "/a")
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusNotFound != resp.StatusCode {
		t.Errorf("expected status to be %v, got %v", http.StatusNotFound, resp.StatusCode)
	}
}

func TestHandler_GetAll(t *testing.T) {
	var store = dummyStore{
		restful.ID("a"): dummyObject{ID: restful.ID("a"), Payload: "hello"},
		restful.ID("b"): dummyObject{ID: restful.ID("b"), Payload: "goodbye"},
	}
	server := newTestServer(store)
	defer server.Close()
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusOK != resp.StatusCode {
		t.Errorf("expected status to be %v, got %v", http.StatusNotFound, resp.StatusCode)
	}
	if resp.Body == nil {
		t.Fatal("expected non nil response body")
	}
	var list []dummyObject
	err = json.NewDecoder(resp.Body).Decode(&list)
	if err != nil {
		t.Fatal(err)
	}
	if len(store) != len(list) {
		t.Errorf("expected len(list) to be %v, got %v", len(store), len(list))
	}
	for i, obj := range list {
		if !reflect.DeepEqual(store[obj.ID], obj) {
			t.Errorf("(#%v) expected obj to be %v, got %v", i, store[obj.ID], obj)
		}
	}
}

func TestHandler_Put(t *testing.T) {
	object := dummyObject{ID: "a", Payload: "hello"}
	methods := []string{"POST", "PUT", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			store := make(dummyStore)
			server := newTestServer(store)
			defer server.Close()
			var b bytes.Buffer
			if err := json.NewEncoder(&b).Encode(&object); err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(method, server.URL, &b)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			t.Logf("req %+v", req)
			resp, err := http.DefaultClient.Do(req)
			t.Logf("resp %+v", resp)
			if err != nil {
				t.Fatal(err)
			}
			if resp.Body == nil {
				t.Fatal("expected non nil response body")
			}
			var o dummyObject
			err = json.NewDecoder(resp.Body).Decode(&o)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(object, o) {
				t.Errorf("expected o to be %v, got %v", object, o)
			}
			if !reflect.DeepEqual(object, store[object.ID]) {
				t.Errorf("expected store[%v] to be %v, got %v", object.ID, object, store[object.ID])
			}
		})
	}
}

func TestHandler_Update(t *testing.T) {
	oldObj := dummyObject{ID: restful.ID("a"), Payload: "hello"}
	newObj := dummyObject{ID: oldObj.ID, Payload: "goodbye"}
	methods := []string{"POST", "PUT", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			store := dummyStore{oldObj.ID: oldObj}
			server := newTestServer(store)
			defer server.Close()
			var b bytes.Buffer
			if err := json.NewEncoder(&b).Encode(&newObj); err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(method, server.URL+"/"+string(newObj.ID), &b)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if http.StatusOK != resp.StatusCode {
				t.Error("expected status code to be %v, got %v", http.StatusOK, resp.StatusCode)
			}
			if !reflect.DeepEqual(newObj, store[newObj.ID]) {
				t.Errorf("expected store[%v] to be %v, got %v", newObj.ID, newObj, store[newObj.ID])
			}
		})
	}
}

func TestHandler_UpdateMissing(t *testing.T) {
	newObj := dummyObject{ID: restful.ID("a"), Payload: "goodbye"}
	methods := []string{"POST", "PUT", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			store := make(dummyStore)
			server := newTestServer(store)
			defer server.Close()
			var b bytes.Buffer
			if err := json.NewEncoder(&b).Encode(&newObj); err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(method, server.URL+"/"+string(newObj.ID), &b)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			if http.StatusNotFound != resp.StatusCode {
				t.Errorf("expected status code to be %v, got %v", http.StatusNotFound, resp.StatusCode)
			}
		})
	}
}

func TestHandler_Delete(t *testing.T) {
	obj := dummyObject{ID: restful.ID("a"), Payload: "hello"}
	store := dummyStore{obj.ID: obj}
	server := newTestServer(store)
	defer server.Close()
	req, err := http.NewRequest("DELETE", server.URL+"/"+string(obj.ID), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusOK != resp.StatusCode {
		t.Error("expected status code to be %v, got %v", http.StatusOK, resp.StatusCode)
	}
}

func TestHandler_DeleteMissing(t *testing.T) {
	store := make(dummyStore)
	server := newTestServer(store)
	defer server.Close()
	req, err := http.NewRequest("DELETE", server.URL+"/a", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if http.StatusNotFound != resp.StatusCode {
		t.Error("expected status code to be %v, got %v", http.StatusNotFound, resp.StatusCode)
	}
}

package restful

import (
	"net/http"
	"strings"
)

// Handler represents a tuple of Codec and Store.
// Handler implements the http.Handler interface.
type Handler struct {
	Codec
	Store
}

// FIXME add logging & customizable error handling
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Store == nil {
		return
	}
	// the trim is so that we don't have to mount the handler at "/<api>" and "/<api>/"
	// when using the default mux
	path := strings.TrimPrefix(r.URL.Path, "/")
	// TODO PATCH support?
	switch r.Method {
	case "GET":
		if path == "" {
			h.handleGetAll(w, r)
		} else {
			h.handleGet(ID(path), w, r)
		}
	case "PUT":
		h.handlePutWithID(ID(path), w, r)
	case "POST":
		h.handlePut(w, r)
	case "DELETE":
		h.handleDelete(ID(path), w, r)
	}
}

func (h Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	list, err := h.GetAll()
	if err != nil {
		// FIXME better error handling
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.Encode(w, list)
	if err != nil {
		// FIXME better error handling
		// FIXME do not expose raw encoding error
		http.Error(w, "", http.StatusInternalServerError)
	}
}

func (h Handler) handleGet(id ID, w http.ResponseWriter, r *http.Request) {
	item, err := h.Get(id)
	if err != nil {
		// FIXME better error handling
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if item == nil {
		// FIXME better error handling
		http.Error(w, "", http.StatusNotFound)
		return
	}
	err = h.Encode(w, item)
	if err != nil {
		// FIXME better error handling
		// FIXME do not expose raw encoding error
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) handlePut(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		// FIXME better error handling
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	item := h.New()
	err := h.Decode(r.Body, item)
	if err != nil {
		// FIXME better error handling
		// FIXME do not expose raw decoding error
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := h.Put(item)
	if err != nil {
		// FIXME better error handling
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// FIXME how to include "id" in model?
	// FIXME should respond with {"id": <id>}
	err = h.Encode(w, id)
	if err != nil {
		// FIXME better error handling
		// FIXME do not encoding raw decoding error
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h Handler) handlePutWithID(id ID, w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		// FIXME better error handling
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	item := h.New()
	err := h.Decode(r.Body, item)
	if err != nil {
		// FIXME better error handling
		// FIXME do not expose raw decoding error
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = h.PutWithId(id, item)
	if err != nil {
		// FIXME better error handling
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// FIXME appropriate response?
}

func (h Handler) handleDelete(id ID, w http.ResponseWriter, r *http.Request) {
	err := h.Delete(id)
	if err != nil {
		// FIXME better error handling
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	// FIXME is there a response to this?
}

// NewJSONHandler creates a new RESTful JSON handler.
func NewJSONHandler(s Store) *Handler {
	return &Handler{
		Codec: JSONCodec,
		Store: s,
	}
}

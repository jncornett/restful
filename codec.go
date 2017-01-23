package restful

import (
	"encoding/json"
	"io"
)

// Codec is an interface that holds the Encode and Decode methods.
type Codec interface {
	Encode(w io.Writer, v interface{}) error
	Decode(r io.Reader, v interface{}) error
}

// ClientCodec is an interface that extends the Codec interface with a
// GetBodyType method.
type ClientCodec interface {
	Codec
	GetBodyType() string
}

type jsonCodec struct{}

func (c jsonCodec) Encode(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

func (c jsonCodec) Decode(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

func (c jsonCodec) GetBodyType() string {
	return "application/json; charset=utf-8"
}

// JSONCodec is the default implementation of ClientCodec.
var JSONCodec ClientCodec = &jsonCodec{}

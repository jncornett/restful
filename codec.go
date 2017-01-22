package restful

import "io"

// Codec is an interface that holds the Encode and Decode methods.
type Codec interface {
	Encode(w io.Writer, v interface{}) error
	Decode(r io.Reader, v interface{}) error
}

// ClientCodec is an interface that extends the Codec interface with a GetBodyType method.
type ClientCodec interface {
	Codec
	GetBodyType() string
}

package resp

import (
	"bytes"
	"fmt"
)

// Array represents an Array response value.
type Array []Value

func (Array) value() {}

func (a Array) writeTo(w *bytes.Buffer) (err error) {
	if a == nil {
		_, err = fmt.Fprintf(w, "*-1\r\n")
		return
	}
	_, err = fmt.Fprintf(w, "*%d\r\n", len(a))
	for i := 0; err == nil && i < len(a); i++ {
		err = a[i].writeTo(w)
	}
	return
}

// BulkString represents a Bulk String response value.
type BulkString []byte

func (s BulkString) writeTo(w *bytes.Buffer) (err error) {
	if s == nil {
		_, err = fmt.Fprintf(w, "$-1\r\n")
		return
	}
	_, err = fmt.Fprintf(w, "$%d\r\n", len(s))
	if err != nil {
		return
	}
	w.Write(s)
	w.WriteString("\r\n")
	return
}

// Error represents an Error response value.
type Error string

func (e Error) Error() string {
	return string(e)
}

func (e Error) writeTo(w *bytes.Buffer) (err error) {
	_, err = fmt.Fprintf(w, "-%s\r\n", e)
	return
}

// Integer represents an Integer response value.
type Integer int64

func (i Integer) writeTo(w *bytes.Buffer) (err error) {
	_, err = fmt.Fprintf(w, ":%d\r\n", i)
	return
}

// SimpleString represents a Simple String response value.
type SimpleString string

func (s SimpleString) String() string {
	return string(s)
}

func (s SimpleString) writeTo(w *bytes.Buffer) (err error) {
	_, err = fmt.Fprintf(w, "+%s\r\n", s)
	return
}

// Value represents any response value.
type Value interface {
	writeTo(*bytes.Buffer) error
}

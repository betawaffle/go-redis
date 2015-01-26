package resp

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
)

const debug = false

type Scanner struct {
	r io.Reader
	s *bufio.Scanner

	skip int
	bulk *BulkStringReader
}

func NewScanner(r io.Reader) *Scanner {
	s := &Scanner{
		r: r,
		s: bufio.NewScanner(r),
	}
	s.s.Split(s.split)
	return s
}

func (s *Scanner) Bytes() []byte {
	return s.s.Bytes()
}

func (s *Scanner) Err() error {
	return s.s.Err()
}

func (s *Scanner) Scan() bool {
	if s.bulk != nil {
		s.bulk.invalidate()
		s.bulk = nil
	}
	return s.s.Scan()
}

func (s *Scanner) Value() interface{} {
	token := s.Bytes()
	switch token[0] {
	case '+':
		return string(token[1:])
	case '-':
		return errors.New(string(token[1:]))
	case ':':
		return parseInt(token[1:])
	case '$':
		return s.bulk
	case '*':
		size := parseInt(token[1:])
		if size == -1 {
			return ArrayHeader(nil)
		}
		return make(ArrayHeader, size)
	}
	return nil
}

func (s *Scanner) dumpValue() {
	switch v := s.Value().(type) {
	case *BulkStringReader:
		b, err := ioutil.ReadAll(v)
		if err != nil {
			log.Print(err)
			return
		}
		fmt.Printf("%q\n", b)
	case ArrayHeader:
		if v == nil {
			fmt.Printf("[](nil)\n")
		} else {

			fmt.Printf("[](%d)\n", len(v))
		}
	default:
		fmt.Printf("%v\n", v)
	}
}

func (s *Scanner) maybeSetupBulk(token, buffered []byte) {
	if token[0] != '$' {
		return
	}

	size := parseInt(token[1:])

	if size == -1 {
		s.bulk = nil
		return
	}
	s.skip = size + 2 // Skip over CRLF

	var r io.Reader
	if len(buffered) >= size {
		r = bytes.NewReader(buffered[:size])
	} else if len(buffered) > 0 {
		r = io.MultiReader(bytes.NewReader(buffered), s.r)
	} else {
		r = s.r
	}
	s.bulk = &BulkStringReader{io.LimitedReader{R: r, N: int64(size)}}
}

func (s *Scanner) split(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if debug {
		defer func() { log.Printf("split(%q, %v) = (%d, %q, %v)", data, atEOF, advance, token, err) }()
	}
	if atEOF && len(data) == 0 {
		return
	}
	if s.skip > 0 {
		if s.skip > len(data) {
			advance = len(data)
			s.skip -= advance
			return
		}
		advance = s.skip
		s.skip -= advance
		data = data[advance:]
	}
again:
	for i, c := range data {
		switch c {
		case '\r':
			token = data[:i]
			data = data[i+1:]

			if data[0] != '\n' {
				err = errors.New("expected newline")
				return
			}
			advance += i + 2
			if i == 0 {
				goto again
			}
			s.maybeSetupBulk(token, data[1:])
			return
		case '\n':
			err = errors.New("unexpected newline")
			return
		}
	}
	return
}

type ArrayHeader []struct{}

type BulkStringReader struct {
	io.LimitedReader
}

func newBulkStringReader(size int, buffered []byte, r io.Reader) *BulkStringReader {
	if size == -1 {
		return nil
	}
	if len(buffered) >= size {
		r = bytes.NewReader(buffered[:size])
	} else if len(buffered) > 0 {
		r = io.MultiReader(bytes.NewReader(buffered), r)
	}
	return &BulkStringReader{io.LimitedReader{R: r, N: int64(size)}}
}

func (b *BulkStringReader) invalidate() {
	b.R = nil
	b.N = -1
}

func parseInt(data []byte) (i int) {
	for _, c := range data {
		i = i*10 + int(c-'0')
	}
	return
}

package rdb

import (
	"bytes"
	"errors"
	"fmt"
	"hash/crc64"
	"io"
	"log"
	"time"
	"unsafe"
)

var (
	magicBytes = [...]byte{'R', 'E', 'D', 'I', 'S'}
	crc64Table = crc64.MakeTable(crc64.ISO)
)

var (
	ErrBadMagic = errors.New("bad magic")
)

type header struct {
	magic   [5]byte
	version [4]byte
}

func (h *header) bytes() []byte {
	return (*[9]byte)(unsafe.Pointer(h))[:]
}

type Parser struct {
	r   io.Reader
	buf bytes.Buffer
	err error
	crc uint64
	vsn uint16
	db  int

	expiry time.Time
}

func NewParser(r io.Reader) *Parser {
	return &Parser{r: r, db: -1}
}

func (p *Parser) Parse() error {
	if err := p.parseHeader(); err != nil {
		return err
	}
	for !p.parse() {
	}
	return p.err
}

func (p *Parser) next(n int) (b []byte, err error) {
	if p.buf.Len() < n {
		p.buf.Grow(n)
		_, err = io.CopyN(&p.buf, p.r, int64(n-p.buf.Len()))
		if err != nil {
			return
		}
	}
	b = p.buf.Next(n)
	p.crc = crc64.Update(p.crc, crc64Table, b)
	fmt.Printf("next: %v\n", b)
	return
}

func (p *Parser) nextByte() (byte, error) {
	b, err := p.next(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (p *Parser) nextLength() (length int, err error) {
	c, err := p.nextByte()
	if err != nil {
		return
	}
	high2, low6 := c>>6, c&63
	switch high2 {
	case 0:
		// 6 bits
		length = int(low6)
	case 1:
		// 14 bits
		c, err = p.nextByte()
		if err != nil {
			return
		}
		length = int(low6)<<8 | int(c)
	case 2:
		// 4 bytes, skip current byte
		var b []byte
		b, err = p.next(4)
		if err != nil {
			return
		}
		length = int(*(*uint32)(unsafe.Pointer(&b[0])))
	case 3:
		// special, 6 bit format
		switch low6 {
		case 0, 1, 2:
			// integers: 8, 16, and 32 bits
			length = -(1 << low6)
		case 4:
			// compressed string
			length = 0
		default:
			err = fmt.Errorf("unexpected special format code: %x", low6)
		}
	}
	return
}

func (p *Parser) parse() bool {
	t, err := p.nextByte()
	if err != nil {
		p.err = err
		return true
	}

	switch t {
	case 0xFF:
		// end of RDB file
		if p.vsn < 5 {
			return true
		}
		// next: CRC 64 checksum of the entire file (8 bytes)
		b, err := p.next(8)
		if err != nil {
			p.err = err
			return true
		}
		switch *(*uint64)(unsafe.Pointer(&b[0])) {
		case 0:
		case p.crc:
		default:
			p.err = errors.New("invalid checksum")
		}
		return true
	case 0xFE:
		// next: database selector (length encoded)
		l, err := p.nextLength()
		if err != nil {
			p.err = err
			return true
		}
		log.Printf("db: %d", l)
		p.db = l
		return false
	case 0xFD:
		// next: expiry time in seconds (4 byte unsigned int)
		b, err := p.next(4)
		if err != nil {
			p.err = err
			return true
		}
		sec := *(*uint32)(unsafe.Pointer(&b[0]))
		p.expiry = time.Unix(int64(sec), 0)
		t, err = p.nextByte()
		if err != nil {
			p.err = err
			return true
		}
	case 0xFC:
		// next: expiry time in ms (8 byte unsigned long)
		b, err := p.next(8)
		if err != nil {
			p.err = err
			return true
		}
		ms := *(*uint64)(unsafe.Pointer(&b[0]))
		p.expiry = unixMilliseconds(ms)
		t, err = p.nextByte()
		if err != nil {
			p.err = err
			return true
		}
	}

	switch t {
	case 0: // string
	case 1: // list
	case 2: // set
	case 3: // sorted set
	case 4: // hash
	case 9: // zipmap
	case 10: // ziplist
	case 11: // intset
	case 12: // sorted set in ziplist
	case 13: // hashmap in ziplist (rdb version 4+)
	default:
		p.err = fmt.Errorf("unexpected value type: %x", t)
		return true
	}
	return false
}

func (p *Parser) parseHeader() error {
	log.Printf("parsing header")
	b, err := p.next(9)
	if err != nil {
		return err
	}
	h := (*header)(unsafe.Pointer(&b[0]))
	log.Printf("header: %s %s", h.magic, h.version)
	if !bytes.Equal(h.magic[:], magicBytes[:]) {
		return ErrBadMagic
	}
	for _, c := range h.version {
		p.vsn = p.vsn*10 + uint16(c)
	}
	return nil
}

func unixMilliseconds(ms uint64) time.Time {
	sec := ms / 1000
	ms -= sec * 1000
	return time.Unix(int64(sec), int64(ms*1000000))
}

package resp

import (
	"bufio"
	"io"
	"log"
)

type Parser struct {
	buf *bufio.Reader
	err error
	val Value
}

func NewParser(r io.Reader) *Parser {
	return &Parser{buf: bufio.NewReader(r)}
}

func (p *Parser) Err() error {
	if p.err == io.EOF {
		return nil
	}
	return p.err
}

func (p *Parser) Parse() bool {
	if p.err != nil {
		return false
	}
	return p.parseAny()
}

func (p *Parser) Serve(replyTo <-chan chan<- Value) error {
	for p.Parse() {
		v := p.Val()
		c := <-replyTo
		select {
		case c <- v:
		default:
			log.Printf("ignoring: %#v", v)
		}
	}
	return p.Err()
}

func (p *Parser) Val() Value {
	return p.val
}

func (p *Parser) parseAny() bool {
	b, err := p.buf.ReadByte()
	if err != nil {
		p.err = err
		return false
	}
	switch b {
	case '+':
		// Simple String
		return p.parseSimpleString()
	case '-':
		// Error
		return p.parseError()
	case ':':
		// Integer
		return p.parseInteger()
	case '$':
		// Bulk String
		return p.parseBulkString()
	case '*':
		// Array
		return p.parseArray()
	default:
		// Inline Command
		// Note: Redis never sends in this format.
		if err := p.buf.UnreadByte(); err != nil {
			p.err = err
			return false
		}
		return p.parseInlineCommand()
	}
}

func (p *Parser) parseArray() bool {
	n := p.readInt()
	if p.err != nil {
		return false
	}
	if n == -1 {
		p.val = Array(nil)
		return true
	}
	arr := make(Array, n)
	for i := range arr {
		if !p.parseAny() {
			return false
		}
		arr[i] = p.val
	}
	p.val = arr
	return true
}

func (p *Parser) parseBulkString() bool {
	n := p.readInt()
	if p.err != nil {
		return false
	}
	if n == -1 {
		p.val = BulkString(nil)
		return true
	}
	str := make(BulkString, n)
	_, err := io.ReadFull(p.buf, str)
	if err != nil {
		p.err = err
		return false
	}
	p.val = str
	return true
}

func (p *Parser) parseError() bool {
	p.val = Error(p.readLine())
	return p.err == nil
}

func (p *Parser) parseInlineCommand() bool {
	line := p.readLine()
	if line == nil {
		return false
	}
	var arr Array
	start := 0
	for i := start; i < len(line); i++ {
		switch line[i] {
		case ' ':
			if i != start {
				arr = append(arr, append(BulkString(nil), line[start:i]...))
				start = i
			}
			start++
		}
	}
	p.val = arr
	return true
}

func (p *Parser) parseInteger() bool {
	p.val = Integer(p.readInt())
	return p.err == nil
}

func (p *Parser) parseSimpleString() bool {
	p.val = SimpleString(p.readLine())
	return p.err == nil
}

func (p *Parser) readInt() (i int64) {
	line := p.readLine()
	if line == nil {
		return 0
	}
	sign := int64(1)
	if line[0] == '-' {
		sign = -1
		line = line[1:]
	}
	for _, c := range line {
		i = i*10 + int64(c-'0')
	}
	return i * sign
}

func (p *Parser) readLine() (buf []byte) {
	b, more, err := p.buf.ReadLine()
	if err != nil || !more {
		p.err = err
		return b
	}

	buf = append(buf, b...)
	for more {
		b, more, err = p.buf.ReadLine()
		if err != nil {
			p.err = err
			return nil
		}
		buf = append(buf, b...)
	}
	return
}

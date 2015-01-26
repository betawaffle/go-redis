package resp

import (
	"bytes"
	"io"
	"net"
)

type Conn struct {
	c net.Conn
	p *Parser
}

// func NewConn(c net.Conn) *Conn {
// }

func (c *Conn) Send(v Value, replyTo chan<- Value) (err error) {
	var buf bytes.Buffer
	err = v.writeTo(&buf)
	if err == nil {
		_, err = io.Copy(c.c, &buf)
	}
	return
}

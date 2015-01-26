package redis

import (
	"bytes"
	"fmt"
	"net"
)

// Conn wraps a net.Conn with Redis functionality.
type Conn struct {
	net.Conn
}

func (c *Conn) WriteCommand(cmd string, args ...string) {
	if cmd == "" {
		panic("empty command")
	}
	var buf bytes.Buffer

	// Array Header
	writeInt(&buf, '*', 1+len(args))

	// Bulk String Header
	writeInt(&buf, '$', len(cmd))

	// Bulk String
	buf.WriteString(cmd)
	buf.WriteString(crlf)

	fmt.Fprintf(&buf, "*%d\r\n$%d\r\n%s\r\n", 1+len(args), len(cmd), cmd)
	for _, arg := range args {
		fmt.Fprintf(&buf, "$%d\r\n%s\r\n", len(arg), arg)
	}
}

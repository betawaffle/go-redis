package rdb

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUnixMilliseconds(t *testing.T) {
	assert.Equal(t, time.Unix(1422118113, 622000000), unixMilliseconds(1422118113622))

	// Ensure it doesn't overflow.
	assert.Equal(t, time.Unix(18446744073709551, 615000000), unixMilliseconds(0xFFFFFFFFFFFFFFFF))
}

func TestSyncLocal(t *testing.T) {
	c, err := syncLocal()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	p := NewParser(c)

	log.Printf("parsing")
	if err := p.Parse(); err != nil {
		panic(err)
	}
}

func syncLocal() (net.Conn, error) {
	c, err := net.Dial("tcp4", "127.0.0.1:6379")
	if err != nil {
		return nil, err
	}
	log.Printf("connected")
	_, err = fmt.Fprintf(c, "SYNC\r\n")
	if err != nil {
		c.Close()
		return nil, err
	}
	var buf [5]byte
	_, err = c.Read(buf[:])
	if err != nil {
		c.Close()
		return nil, err
	}
	return c, nil
}

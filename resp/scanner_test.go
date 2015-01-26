package resp

import (
	"fmt"

	"net"
	"runtime"
	"testing"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func TestLocal(t *testing.T) {
	c, err := net.Dial("tcp4", "127.0.0.1:6379")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	s := NewScanner(c)

	go exerciseRedis(c)

	for s.Scan() {
		s.dumpValue()
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
}

func exerciseRedis(c net.Conn) {
	fmt.Fprintf(c, "SET foo bar\r\n")
	fmt.Fprintf(c, "GET foo\r\n")
	fmt.Fprintf(c, "DEL foo\r\n")
	fmt.Fprintf(c, "SADD foo bar baz 1\r\n")
	fmt.Fprintf(c, "SMEMBERS foo\r\n")
	fmt.Fprintf(c, "DEL foo\r\n")
	// time.Sleep(5 * time.Second)
	fmt.Fprintf(c, "QUIT\r\n")
}

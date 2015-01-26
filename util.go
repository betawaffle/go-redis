package redis

import "bytes"

const crlf = "\r\n"

const (
	maxLen = 512 * 1024 * 1024 // 512 MB
	maxBuf = 1 + len(string(maxLen)) + len(crlf) + maxLen + len(crlf)
)

func writeGET(w *bytes.Buffer, key []byte) {
	const (
		cmd = "GET"
		arr = "*" + string(2) + crlf + "$" + string(len(cmd)) + crlf + cmd + crlf
	)
	w.Grow(len(arr) + len(key) + len(crlf))
	w.WriteString(arr)
	w.Write(key)
	w.WriteString(crlf)
}

func writeBulkBytes(w *bytes.Buffer, str []byte) {
	writeInt(w, '$', len(str))
	w.Write(str)
	w.WriteString(crlf)
}

func writeBulkString(w *bytes.Buffer, str string) {
	writeInt(w, '$', len(str))
	w.WriteString(str)
	w.WriteString(crlf)
}

func writeInt(w *bytes.Buffer, t byte, v int) {
	// 1 byte for the type character (':' or '$' or '*').
	// 20 bytes for worst-case signed 64-bit integer in decimal.
	// 2 bytes for the CRLF.
	const bufLen = 1 + 20 + len(crlf)

	var (
		buf [bufLen]byte
		neg = v < 0
		i   = len(buf) - 2
	)

	// Write the CRLF.
	buf[i+0] = crlf[0]
	buf[i+1] = crlf[1]

	// Write all the digits.
	if neg {
		v = -v
	}
	uv := uint64(v)
	for uv >= 10 {
		i--
		next := uv / 10
		buf[i] = byte('0' + uv - next*10)
		uv = next
	}
	i--
	buf[i] = byte('0' + uv)

	// Write the sign, if any.
	if neg {
		i--
		buf[i] = '-'
	}

	// Write the type character.
	i--
	buf[i] = t

	w.Write(buf[i:])
}

// type StreamConn interface {
// 	Close() error
// 	CloseRead() error
// 	CloseWrite() error
// 	File() (f *os.File, err error)
// 	LocalAddr() net.Addr
// 	Read(b []byte) (int, error)
// 	RemoteAddr() net.Addr
// 	SetDeadline(t time.Time) error
// 	SetReadBuffer(bytes int) error
// 	SetReadDeadline(t time.Time) error
// 	SetWriteBuffer(bytes int) error
// 	SetWriteDeadline(t time.Time) error
// 	Write(b []byte) (int, error)
// }

package commands

// WriteAppend writes an APPEND command to the buffer.
func (b *Buffer) WriteAppend(key, value []byte) {
	const cmd = "*3" + crlf + "$6" + crlf + "APPEND" + crlf
	b.WriteString(cmd)
	b.WriteBulkBytes(key)
	b.WriteBulkBytes(value)
}

// WriteBitCount writes a BITCOUNT command to the buffer.
func (b *Buffer) WriteBitCount(key []byte) {
	const cmd = "*2" + crlf + "$8" + crlf + "BITCOUNT" + crlf
	b.WriteString(cmd)
	b.WriteBulkBytes(key)
}

// WriteBitCountRange writes a BITCOUNT command to the buffer.
func (b *Buffer) WriteBitCountRange(key []byte, start, end int) {
	const cmd = "*4" + crlf + "$8" + crlf + "BITCOUNT" + crlf
	b.WriteString(cmd)
	b.WriteBulkBytes(key)
	b.WriteBulkInt(start)
	b.WriteBulkInt(end)
}

// WriteBitOp writes a BITOP command to the buffer.
func (b *Buffer) WriteBitOp(operation string, destkey []byte, keys ...[]byte) {
	b.WriteCommand("BITOP", 2+len(keys))
	b.WriteBulkString(operation)
	b.WriteBulkBytes(destkey)
	for _, key := range keys {
		b.WriteBulkBytes(key)
	}
}

// WriteBitPos writes a BITPOS command to the buffer.
func (b *Buffer) WriteBitPos(key []byte, set bool) {
	const cmd = "*3" + crlf + "$6" + crlf + "BITPOS" + crlf
	b.WriteString(cmd)
	b.WriteBulkBytes(key)
	if set {
		b.writeInt(1, true)
	} else {
		b.writeInt(0, true)
	}
}

// WriteBitPosRange writes a BITPOS command to the buffer.
func (b *Buffer) WriteBitPosRange(key []byte, set bool, start, end int) {
	const cmd = "*5" + crlf + "$6" + crlf + "BITPOS" + crlf
	b.WriteString(cmd)
	b.WriteBulkBytes(key)
	if set {
		b.writeInt(1, true)
	} else {
		b.writeInt(0, true)
	}
	b.writeInt(start, true)
	b.writeInt(end, true)
}

// WriteGet writes a GET command to the buffer.
func (b *Buffer) WriteGet(key []byte) {
	const cmd = "*2" + crlf + "$3" + crlf + "GET" + crlf
	b.WriteString(cmd)
	b.WriteBulkBytes(key)
}

// WriteMultiGet writes an MGET command to the buffer.
func (b *Buffer) WriteMultiGet(keys ...[]byte) {
	b.WriteCommand("MGET", len(keys))
	for _, key := range keys {
		b.WriteBulkBytes(key)
	}
}

// func WriteGET(w Buffer, key []byte) {
// 	const (
// 		cmd = "GET"
// 		arr = "*" + string(2) + crlf + "$" + string(len(cmd)) + crlf + cmd + crlf
// 	)
// 	w.Grow(len(arr) + len(key) + len(crlf))
// 	w.WriteString(arr)
// 	w.Write(key)
// 	w.WriteString(crlf)
// }

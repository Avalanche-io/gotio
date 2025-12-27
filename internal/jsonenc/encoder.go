// SPDX-License-Identifier: Apache-2.0
// Copyright Contributors to the OpenTimelineIO project

// Package jsonenc provides a high-performance streaming JSON encoder
// optimized for OTIO serialization.
package jsonenc

import (
	"io"
	"math"
	"strconv"
	"sync"
	"unicode/utf8"
)

// Encoder is a streaming JSON encoder that writes directly to an io.Writer.
// It avoids intermediate allocations by building JSON incrementally.
type Encoder struct {
	w         io.Writer
	buf       []byte
	scratch   [64]byte
	err       error
	needComma bool
}

// bufferPool provides reusable buffers for encoders
var bufferPool = sync.Pool{
	New: func() any {
		return make([]byte, 0, 4096)
	},
}

// NewEncoder creates a new streaming encoder writing to w.
func NewEncoder(w io.Writer) *Encoder {
	buf := bufferPool.Get().([]byte)
	return &Encoder{
		w:   w,
		buf: buf[:0],
	}
}

// Release returns the encoder's buffer to the pool.
// The encoder should not be used after calling Release.
func (e *Encoder) Release() {
	if e.buf != nil {
		bufferPool.Put(e.buf[:0])
		e.buf = nil
	}
}

// Flush writes any buffered data to the underlying writer.
func (e *Encoder) Flush() error {
	if e.err != nil {
		return e.err
	}
	if len(e.buf) > 0 {
		_, e.err = e.w.Write(e.buf)
		e.buf = e.buf[:0]
	}
	return e.err
}

// Error returns any error that occurred during encoding.
func (e *Encoder) Error() error {
	return e.err
}

// Bytes returns the encoded JSON as bytes.
// Only valid if the encoder was created with a bytes.Buffer.
func (e *Encoder) Bytes() []byte {
	return e.buf
}

// grow ensures the buffer has space for n more bytes
func (e *Encoder) grow(n int) {
	if cap(e.buf)-len(e.buf) < n {
		newCap := cap(e.buf) * 2
		if newCap < len(e.buf)+n {
			newCap = len(e.buf) + n
		}
		newBuf := make([]byte, len(e.buf), newCap)
		copy(newBuf, e.buf)
		e.buf = newBuf
	}
}

// writeByte writes a single byte
func (e *Encoder) writeByte(b byte) {
	e.grow(1)
	e.buf = append(e.buf, b)
}

// writeBytes writes multiple bytes
func (e *Encoder) writeBytes(b []byte) {
	e.grow(len(b))
	e.buf = append(e.buf, b...)
}

// writeString writes a string without quotes
func (e *Encoder) writeString(s string) {
	e.grow(len(s))
	e.buf = append(e.buf, s...)
}

// BeginObject writes '{' and resets comma state
func (e *Encoder) BeginObject() {
	e.writeByte('{')
	e.needComma = false
}

// EndObject writes '}'
func (e *Encoder) EndObject() {
	e.writeByte('}')
	e.needComma = true
}

// BeginArray writes '[' and resets comma state
func (e *Encoder) BeginArray() {
	e.writeByte('[')
	e.needComma = false
}

// EndArray writes ']'
func (e *Encoder) EndArray() {
	e.writeByte(']')
	e.needComma = true
}

// WriteComma writes ',' if needed for array/object separation
func (e *Encoder) WriteComma() {
	if e.needComma {
		e.writeByte(',')
	}
	e.needComma = true
}

// WriteKey writes a quoted object key followed by ':'
// Automatically handles comma before the key
func (e *Encoder) WriteKey(key string) {
	if e.needComma {
		e.writeByte(',')
	}
	e.WriteQuotedString(key)
	e.writeByte(':')
	e.needComma = false
}

// WriteNull writes 'null'
func (e *Encoder) WriteNull() {
	e.writeString("null")
	e.needComma = true
}

// WriteBool writes 'true' or 'false'
func (e *Encoder) WriteBool(v bool) {
	if v {
		e.writeString("true")
	} else {
		e.writeString("false")
	}
	e.needComma = true
}

// WriteInt writes an integer value
func (e *Encoder) WriteInt(v int) {
	e.writeString(strconv.Itoa(v))
	e.needComma = true
}

// WriteInt64 writes a 64-bit integer value
func (e *Encoder) WriteInt64(v int64) {
	b := strconv.AppendInt(e.scratch[:0], v, 10)
	e.writeBytes(b)
	e.needComma = true
}

// WriteFloat64 writes a float64 value.
// Handles special values (Inf, NaN) for Python compatibility.
func (e *Encoder) WriteFloat64(v float64) {
	e.WriteFloat64WithOptions(v, true)
}

// WriteFloat64WithOptions writes a float64 with configurable Inf/NaN handling.
// If allowSpecial is true, writes Python-compatible Infinity/NaN literals.
// If false, writes null for special values.
func (e *Encoder) WriteFloat64WithOptions(v float64, allowSpecial bool) {
	if math.IsNaN(v) {
		if allowSpecial {
			e.writeString("NaN")
		} else {
			e.writeString("null")
		}
		e.needComma = true
		return
	}
	if math.IsInf(v, 1) {
		if allowSpecial {
			e.writeString("Infinity")
		} else {
			e.writeString("null")
		}
		e.needComma = true
		return
	}
	if math.IsInf(v, -1) {
		if allowSpecial {
			e.writeString("-Infinity")
		} else {
			e.writeString("null")
		}
		e.needComma = true
		return
	}

	// Use strconv for normal floats
	b := strconv.AppendFloat(e.scratch[:0], v, 'g', -1, 64)
	e.writeBytes(b)
	e.needComma = true
}

// WriteQuotedString writes a JSON-escaped quoted string
func (e *Encoder) WriteQuotedString(s string) {
	e.writeByte('"')
	e.writeEscapedString(s)
	e.writeByte('"')
	e.needComma = true
}

// writeEscapedString writes string contents with JSON escaping
func (e *Encoder) writeEscapedString(s string) {
	start := 0
	for i := 0; i < len(s); {
		b := s[i]
		if b < utf8.RuneSelf {
			if htmlSafeSet[b] {
				i++
				continue
			}
			if start < i {
				e.writeString(s[start:i])
			}
			e.writeByte('\\')
			switch b {
			case '"', '\\':
				e.writeByte(b)
			case '\b':
				e.writeByte('b')
			case '\f':
				e.writeByte('f')
			case '\n':
				e.writeByte('n')
			case '\r':
				e.writeByte('r')
			case '\t':
				e.writeByte('t')
			default:
				// Unicode escape
				e.writeString("u00")
				e.writeByte(hex[b>>4])
				e.writeByte(hex[b&0xF])
			}
			i++
			start = i
			continue
		}
		// Multi-byte UTF-8
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size == 1 {
			if start < i {
				e.writeString(s[start:i])
			}
			e.writeString(`\ufffd`)
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(s) {
		e.writeString(s[start:])
	}
}

// WriteRawJSON writes pre-encoded JSON bytes directly
func (e *Encoder) WriteRawJSON(data []byte) {
	e.writeBytes(data)
	e.needComma = true
}

// WriteStringField writes a key-value pair where value is a string
func (e *Encoder) WriteStringField(key, value string) {
	e.WriteKey(key)
	e.WriteQuotedString(value)
}

// WriteBoolField writes a key-value pair where value is a bool
func (e *Encoder) WriteBoolField(key string, value bool) {
	e.WriteKey(key)
	e.WriteBool(value)
}

// WriteIntField writes a key-value pair where value is an int
func (e *Encoder) WriteIntField(key string, value int) {
	e.WriteKey(key)
	e.WriteInt(value)
}

// WriteFloat64Field writes a key-value pair where value is a float64
func (e *Encoder) WriteFloat64Field(key string, value float64) {
	e.WriteKey(key)
	e.WriteFloat64(value)
}

// WriteNullField writes a key-value pair where value is null
func (e *Encoder) WriteNullField(key string) {
	e.WriteKey(key)
	e.WriteNull()
}

var hex = "0123456789abcdef"

// htmlSafeSet marks characters that don't need escaping
var htmlSafeSet = [utf8.RuneSelf]bool{
	' ':      true,
	'!':      true,
	'"':      false,
	'#':      true,
	'$':      true,
	'%':      true,
	'&':      true,
	'\'':     true,
	'(':      true,
	')':      true,
	'*':      true,
	'+':      true,
	',':      true,
	'-':      true,
	'.':      true,
	'/':      true,
	'0':      true,
	'1':      true,
	'2':      true,
	'3':      true,
	'4':      true,
	'5':      true,
	'6':      true,
	'7':      true,
	'8':      true,
	'9':      true,
	':':      true,
	';':      true,
	'<':      true,
	'=':      true,
	'>':      true,
	'?':      true,
	'@':      true,
	'A':      true,
	'B':      true,
	'C':      true,
	'D':      true,
	'E':      true,
	'F':      true,
	'G':      true,
	'H':      true,
	'I':      true,
	'J':      true,
	'K':      true,
	'L':      true,
	'M':      true,
	'N':      true,
	'O':      true,
	'P':      true,
	'Q':      true,
	'R':      true,
	'S':      true,
	'T':      true,
	'U':      true,
	'V':      true,
	'W':      true,
	'X':      true,
	'Y':      true,
	'Z':      true,
	'[':      true,
	'\\':     false,
	']':      true,
	'^':      true,
	'_':      true,
	'`':      true,
	'a':      true,
	'b':      true,
	'c':      true,
	'd':      true,
	'e':      true,
	'f':      true,
	'g':      true,
	'h':      true,
	'i':      true,
	'j':      true,
	'k':      true,
	'l':      true,
	'm':      true,
	'n':      true,
	'o':      true,
	'p':      true,
	'q':      true,
	'r':      true,
	's':      true,
	't':      true,
	'u':      true,
	'v':      true,
	'w':      true,
	'x':      true,
	'y':      true,
	'z':      true,
	'{':      true,
	'|':      true,
	'}':      true,
	'~':      true,
	'\u007f': true,
}

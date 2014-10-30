// Go support for Protocol Buffers - Google's data interchange format
//
// Copyright 2010 The Go Authors.  All rights reserved.
// http://code.google.com/p/goprotobuf/
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//     * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//     * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package protobufraw

import (
	"errors"
	"fmt"
	"io"
)

// code taken from code.google.com/p/protobuf/

// Constants that identify the encoding of a value on the wire.
const (
	WireVarint     = 0
	WireFixed64    = 1
	WireBytes      = 2
	WireStartGroup = 3
	WireEndGroup   = 4
	WireFixed32    = 5
)

// errOverflow is returned when an integer is too large to be represented.
var errOverflow = errors.New("proto: integer overflow")

type Buffer struct {
	buf   []byte
	index int
}

// DecodeVarint reads a varint-encoded integer from the slice.
// It returns the integer and the number of bytes consumed, or
// zero if there is not enough.
// This is the format for the
// int32, int64, uint32, uint64, bool, and enum
// protocol buffer types.
func DecodeVarint(buf []byte) (x uint64, n int) {
	// x, n already 0
	for shift := uint(0); shift < 64; shift += 7 {
		if n >= len(buf) {
			return 0, 0
		}
		b := uint64(buf[n])
		n++
		x |= (b & 0x7F) << shift
		if (b & 0x80) == 0 {
			return x, n
		}
	}

	// The number is too large to represent in a 64-bit value.
	return 0, 0
}

// DecodeVarint reads a varint-encoded integer from the Buffer.
// This is the format for the
// int32, int64, uint32, uint64, bool, and enum
// protocol buffer types.
func (p *Buffer) DecodeVarint() (x uint64, err error) {
	// x, err already 0

	i := p.index
	l := len(p.buf)

	for shift := uint(0); shift < 64; shift += 7 {
		if i >= l {
			err = io.ErrUnexpectedEOF
			return
		}
		b := p.buf[i]
		i++
		x |= (uint64(b) & 0x7F) << shift
		if b < 0x80 {
			p.index = i
			return
		}
	}

	// The number is too large to represent in a 64-bit value.
	err = errOverflow
	return
}

// DecodeFixed64 reads a 64-bit integer from the Buffer.
// This is the format for the
// fixed64, sfixed64, and double protocol buffer types.
func (p *Buffer) DecodeFixed64() (x uint64, err error) {
	// x, err already 0
	i := p.index + 8
	if i < 0 || i > len(p.buf) {
		err = io.ErrUnexpectedEOF
		return
	}
	p.index = i

	x = uint64(p.buf[i-8])
	x |= uint64(p.buf[i-7]) << 8
	x |= uint64(p.buf[i-6]) << 16
	x |= uint64(p.buf[i-5]) << 24
	x |= uint64(p.buf[i-4]) << 32
	x |= uint64(p.buf[i-3]) << 40
	x |= uint64(p.buf[i-2]) << 48
	x |= uint64(p.buf[i-1]) << 56
	return
}

// DecodeFixed32 reads a 32-bit integer from the Buffer.
// This is the format for the
// fixed32, sfixed32, and float protocol buffer types.
func (p *Buffer) DecodeFixed32() (x uint64, err error) {
	// x, err already 0
	i := p.index + 4
	if i < 0 || i > len(p.buf) {
		err = io.ErrUnexpectedEOF
		return
	}
	p.index = i

	x = uint64(p.buf[i-4])
	x |= uint64(p.buf[i-3]) << 8
	x |= uint64(p.buf[i-2]) << 16
	x |= uint64(p.buf[i-1]) << 24
	return
}

// DecodeZigzag64 reads a zigzag-encoded 64-bit integer
// from the Buffer.
// This is the format used for the sint64 protocol buffer type.
func (p *Buffer) DecodeZigzag64() (x uint64, err error) {
	x, err = p.DecodeVarint()
	if err != nil {
		return
	}
	x = (x >> 1) ^ uint64((int64(x&1)<<63)>>63)
	return
}

// DecodeZigzag32 reads a zigzag-encoded 32-bit integer
// from  the Buffer.
// This is the format used for the sint32 protocol buffer type.
func (p *Buffer) DecodeZigzag32() (x uint64, err error) {
	x, err = p.DecodeVarint()
	if err != nil {
		return
	}
	x = uint64((uint32(x) >> 1) ^ uint32((int32(x&1)<<31)>>31))
	return
}

// These are not ValueDecoders: they produce an array of bytes or a string.
// bytes, embedded messages

// DecodeRawBytes reads a count-delimited byte buffer from the Buffer.
// This is the format used for the bytes protocol buffer
// type and for embedded messages.
func (p *Buffer) DecodeRawBytes(alloc bool) (buf []byte, err error) {
	n, err := p.DecodeVarint()
	if err != nil {
		return
	}

	nb := int(n)
	if nb < 0 {
		return nil, fmt.Errorf("proto: bad byte length %d", nb)
	}
	end := p.index + nb
	if end < p.index || end > len(p.buf) {
		return nil, io.ErrUnexpectedEOF
	}

	if !alloc {
		// todo: check if can get more uses of alloc=false
		buf = p.buf[p.index:end]
		p.index += nb
		return
	}

	buf = make([]byte, nb)
	copy(buf, p.buf[p.index:])
	p.index += nb
	return
}

// DecodeStringBytes reads an encoded string from the Buffer.
// This is the format used for the proto2 string type.
func (p *Buffer) DecodeStringBytes() (s string, err error) {
	buf, err := p.DecodeRawBytes(false)
	if err != nil {
		return
	}
	return string(buf), nil
}

// starts code written by me

var (
	debugLogging = true
)

var wireTypeNames = []string{
	"WireVarint",
	"WireFixed64",
	"WireBytes",
	"WireStartGroup",
	"WireEndGroup",
	"WireFixed32",
}

func WireTypeToString(wire int) string {
	if wire < 0 || wire >= len(wireTypeNames) {
		return "WireInvalid"
	}
	return wireTypeNames[wire]
}

type Field struct {
	Tag          int
	WireType     int // Wire* types
	WireTypeName string
	Offset       int
	Len          int
	Value        interface{} // depend of Type, TODO: add accessor functions
}

func dbglogf(format string, v ...interface{}) {
	if !debugLogging {
		return
	}
	fmt.Printf(format, v...)
}

func (o *Buffer) DecodeRawRecur(groupNest int) ([]Field, error) {
	var err error
	fields := make([]Field, 0)
	for err == nil && o.index < len(o.buf) {
		var field Field
		field.Offset = o.index

		var u uint64
		u, err = o.DecodeVarint()
		if err != nil {
			break
		}
		wire := int(u & 0x7)
		field.WireType = wire
		field.WireTypeName = WireTypeToString(wire)
		if wire == WireEndGroup {
			groupNest--
			dbglogf("WireEndGroup, nest: %d\n", groupNest)
			if groupNest < 0 {
				err = errors.New("proto: group nesting falls below 0")
			}
		}
		tag := int(u >> 3)
		if tag <= 0 {
			return fields, fmt.Errorf("proto: illegal tag %d (wire type %d)", tag, wire)
		}
		field.Tag = tag

		switch wire {
		case WireVarint:
			field.Value, err = o.DecodeVarint()
			if err == nil {
				field.Len = o.index - field.Offset
				fields = append(fields, field)
				dbglogf("idx: %d, len: %d, wire: %s, tag: %d, val: %d\n", field.Offset, field.Len, field.WireTypeName, field.Tag, field.Value)
			} else {
				dbglogf("idx: %d, wire: %s, tag: %d, err: %s\n", field.Offset, field.WireTypeName, field.Tag, err)
			}
		case WireFixed64:
			field.Value, err = o.DecodeFixed64()
			if err == nil {
				field.Len = o.index - field.Offset
				fields = append(fields, field)
				dbglogf("idx: %d, len: %d, wire: %s, tag: %d, val: %d\n", field.Offset, field.Len, field.WireTypeName, field.Tag, field.Value)
			} else {
				dbglogf("idx: %d, wire: %s, tag: %d, err: %s\n", field.Offset, field.WireTypeName, field.Tag, err)
			}
		case WireBytes:
			// TODO: this can be embedded field. try to decode field.Value as a separate buffer
			field.Value, err = o.DecodeRawBytes(false)
			if err == nil {
				field.Len = o.index - field.Offset
				fields = append(fields, field)
				dbglogf("idx: %d, len: %d, wire: %s, tag: %d, val: %v\n", field.Offset, field.Len, field.WireTypeName, field.Tag, field.Value)
			} else {
				dbglogf("idx: %d, wire: %s, tag: %d, err: %s\n", field.Offset, field.WireTypeName, field.Tag, err)
			}
		case WireFixed32:
			field.Value, err = o.DecodeFixed32()
			if err == nil {
				field.Len = o.index - field.Offset
				fields = append(fields, field)
				dbglogf("idx: %d, len: %d, wire: %s, tag: %d, val: %d\n", field.Offset, field.Len, field.WireTypeName, field.Tag, field.Value)
			} else {
				dbglogf("idx: %d, wire: %s, tag: %d, err: %s\n", field.Offset, field.WireTypeName, field.Tag, err)
			}
		case WireStartGroup:
			groupNest++
			dbglogf("enter WireStartGroup, nest: %d\n", groupNest)
			field.Value, err = o.DecodeRawRecur(groupNest)
			if err == nil {
				field.Len = o.index - field.Offset
				fields = append(fields, field)
				dbglogf("idx: %d, len: %d, wire: %s, tag: %d\n", field.Offset, field.Len, field.WireTypeName, field.Tag)
			} else {
				dbglogf("idx: %d, wire: %s, tag: %d, err: %s\n", field.Offset, field.WireTypeName, field.Tag, err)
			}
		default:
			err = fmt.Errorf("proto: can't decode unknown wire type %d", wire)
		}
	}
	return fields, err
}

func DecodeRaw(d []byte) ([]Field, error) {
	o := &Buffer{buf: d, index: 0}
	return o.DecodeRawRecur(0)
}

// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gob

import (
	"bytes"
	"io"
	"os"
	"reflect"
	"sync"
)

// A Decoder manages the receipt of type and data information read from the
// remote side of a connection.
type Decoder struct {
	mutex        sync.Mutex                              // each item must be received atomically
	r            io.Reader                               // source of the data
	wireType     map[typeId]*wireType                    // map from remote ID to local description
	decoderCache map[reflect.Type]map[typeId]**decEngine // cache of compiled engines
	ignorerCache map[typeId]**decEngine                  // ditto for ignored objects
	state        *decodeState                            // reads data from in-memory buffer
	countState   *decodeState                            // reads counts from wire
	buf          []byte
	countBuf     [9]byte // counts may be uint64s (unlikely!), require 9 bytes
}

// NewDecoder returns a new decoder that reads from the io.Reader.
func NewDecoder(r io.Reader) *Decoder {
	dec := new(Decoder)
	dec.r = r
	dec.wireType = make(map[typeId]*wireType)
	dec.state = newDecodeState(nil) // buffer set in Decode(); rest is unimportant
	dec.decoderCache = make(map[reflect.Type]map[typeId]**decEngine)
	dec.ignorerCache = make(map[typeId]**decEngine)

	return dec
}

func (dec *Decoder) recvType(id typeId) {
	// Have we already seen this type?  That's an error
	if dec.wireType[id] != nil {
		dec.state.err = os.ErrorString("gob: duplicate type received")
		return
	}

	// Type:
	wire := new(wireType)
	dec.state.err = dec.decode(tWireType, reflect.NewValue(wire))
	// Remember we've seen this type.
	dec.wireType[id] = wire
}

// Decode reads the next value from the connection and stores
// it in the data represented by the empty interface value.
// The value underlying e must be the correct type for the next
// data item received, and must be a pointer.
func (dec *Decoder) Decode(e interface{}) os.Error {
	value := reflect.NewValue(e)
	// If e represents a value as opposed to a pointer, the answer won't
	// get back to the caller.  Make sure it's a pointer.
	if value.Type().Kind() != reflect.Ptr {
		dec.state.err = os.ErrorString("gob: attempt to decode into a non-pointer")
		return dec.state.err
	}
	return dec.DecodeValue(value)
}

// DecodeValue reads the next value from the connection and stores
// it in the data represented by the reflection value.
// The value must be the correct type for the next
// data item received.
func (dec *Decoder) DecodeValue(value reflect.Value) os.Error {
	// Make sure we're single-threaded through here.
	dec.mutex.Lock()
	defer dec.mutex.Unlock()

	dec.state.err = nil
	for {
		// Read a count.
		var nbytes uint64
		nbytes, dec.state.err = decodeUintReader(dec.r, dec.countBuf[0:])
		if dec.state.err != nil {
			break
		}
		// Allocate the buffer.
		if nbytes > uint64(len(dec.buf)) {
			dec.buf = make([]byte, nbytes+1000)
		}
		dec.state.b = bytes.NewBuffer(dec.buf[0:nbytes])

		// Read the data
		_, dec.state.err = io.ReadFull(dec.r, dec.buf[0:nbytes])
		if dec.state.err != nil {
			if dec.state.err == os.EOF {
				dec.state.err = io.ErrUnexpectedEOF
			}
			break
		}

		// Receive a type id.
		id := typeId(decodeInt(dec.state))
		if dec.state.err != nil {
			break
		}

		// Is it a new type?
		if id < 0 { // 0 is the error state, handled above
			// If the id is negative, we have a type.
			dec.recvType(-id)
			if dec.state.err != nil {
				break
			}
			continue
		}

		// No, it's a value.
		// Make sure the type has been defined already or is a builtin type (for
		// top-level singleton values).
		if dec.wireType[id] == nil && builtinIdToType[id] == nil {
			dec.state.err = errBadType
			break
		}
		dec.state.err = dec.decode(id, value)
		break
	}
	return dec.state.err
}

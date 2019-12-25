// Package prefbuf allocates buffers that allow for cheaply prefixing data.
package prefbuf

import (
	"unsafe"
)

type buffer struct {
	buf   []byte
	extra int
}

const Extra = 32

var active = map[uintptr]*buffer{}

func Alloc(n int) []byte {
	larger := make([]byte, n+Extra)
	ret := larger[Extra:]
	active[ptr(ret)] = &buffer{larger, Extra}
	return ret
}

func Free(b []byte) {
	_, ok := active[ptr(b)]
	if !ok {
		panic("double free")
	}
	delete(active, ptr(b))
}

func Prefix(prefix, b []byte) []byte {
	bfr, ok := active[ptr(b)]
	if !ok {
		panic("Prefix called on non-active buffer")
	}
	delete(active, ptr(b))
	bfr.extra -= len(prefix)
	ret := bfr.buf[bfr.extra:]
	active[ptr(ret)] = bfr
	copy(ret, prefix)
	return ret
}

func Unprefix(b []byte, prefLen int) {
	bfr, ok := active[ptr(b)]
	if !ok {
		panic("Unprefix called on non-active buffer")
	}
	delete(active, ptr(b))
	bfr.extra += prefLen
	active[ptr(bfr.buf[bfr.extra:])] = bfr
}

func ptr(b []byte) uintptr {
	return uintptr(unsafe.Pointer(&b[0]))
}

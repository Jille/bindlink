package prefbuf_test

import (
	"strings"
	"testing"

	"github.com/Jille/bindlink/prefbuf"
)

func TestPrefBuf(t *testing.T) {
	b := prefbuf.Alloc(10)
	copy(b, strings.Repeat("x", 10))
	if string(b) != "xxxxxxxxxx" {
		t.Fatalf("b is %q, expected xxxxxxxxxx", string(b))
	}
	pb := prefbuf.Prefix([]byte("yyy"), b)
	if string(pb) != "yyyxxxxxxxxxx" {
		t.Fatalf("pb is %q, expected yyyxxxxxxxxxx", string(pb))
	}
	pb2 := prefbuf.Prefix([]byte("zzz"), pb)
	if string(pb2) != "zzzyyyxxxxxxxxxx" {
		t.Fatalf("pb2 is %q, expected zzzyyyxxxxxxxxxx", string(pb2))
	}
	prefbuf.Unprefix(pb2, 3)
	prefbuf.Unprefix(pb, 3)
	pb = prefbuf.Prefix([]byte("aaa"), b)
	if string(pb) != "aaaxxxxxxxxxx" {
		t.Fatalf("pb is %q, expected aaaxxxxxxxxxx", string(pb))
	}
	prefbuf.Unprefix(pb, 3)
	prefbuf.Free(b)
}

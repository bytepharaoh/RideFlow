package middleware

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"testing"
)

type hijackFlushWriter struct {
	header      http.Header
	flushed     bool
	hijackCalls int
}

func (w *hijackFlushWriter) Header() http.Header {
	return w.header
}

func (w *hijackFlushWriter) Write(p []byte) (int, error) {
	return len(p), nil
}

func (w *hijackFlushWriter) WriteHeader(int) {}

func (w *hijackFlushWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	w.hijackCalls++
	return nil, bufio.NewReadWriter(bufio.NewReader(bytes.NewReader(nil)), bufio.NewWriter(io.Discard)), nil
}

func (w *hijackFlushWriter) Flush() {
	w.flushed = true
}

func TestResponseWriterHijackDelegates(t *testing.T) {
	t.Parallel()

	base := &hijackFlushWriter{header: make(http.Header)}
	wrapped := NewResponseHandler(base)

	_, _, err := wrapped.Hijack()
	if err != nil {
		t.Fatalf("Hijack() error = %v", err)
	}

	if base.hijackCalls != 1 {
		t.Fatalf("hijack calls = %d, want %d", base.hijackCalls, 1)
	}
}

func TestResponseWriterFlushDelegates(t *testing.T) {
	t.Parallel()

	base := &hijackFlushWriter{header: make(http.Header)}
	wrapped := NewResponseHandler(base)

	wrapped.Flush()

	if !base.flushed {
		t.Fatal("expected Flush() to delegate to underlying ResponseWriter")
	}
}

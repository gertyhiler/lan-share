package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

type flushRecorder struct {
	*httptest.ResponseRecorder
	flushed bool
}

func (r *flushRecorder) Flush() {
	r.flushed = true
}

func TestLogResponseWriterPreservesFlush(t *testing.T) {
	t.Parallel()
	rec := &flushRecorder{ResponseRecorder: httptest.NewRecorder()}
	lrw := &logResponseWriter{ResponseWriter: rec}

	flusher, ok := any(lrw).(http.Flusher)
	if !ok {
		t.Fatalf("logResponseWriter does not implement http.Flusher")
	}
	flusher.Flush()
	if !rec.flushed {
		t.Fatalf("flush was not delegated")
	}
}

func TestLogResponseWriterTracksStatus(t *testing.T) {
	t.Parallel()
	rec := httptest.NewRecorder()
	lrw := &logResponseWriter{ResponseWriter: rec}

	lrw.WriteHeader(http.StatusCreated)

	if lrw.code != http.StatusCreated {
		t.Fatalf("code = %d, want %d", lrw.code, http.StatusCreated)
	}
}

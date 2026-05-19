package httpadapter

import (
	"bytes"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexTemplateExecute(t *testing.T) {
	t.Parallel()
	tpl, err := template.ParseFS(indexFS, "templates/index.html")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var buf bytes.Buffer
	err = tpl.Execute(&buf, indexData{
		ShareHost:  "192.168.1.10",
		Port:       8000,
		UploadsDir: "/tmp/uploads",
		SharedDir:  "/tmp/shared",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `id="messages"`) || !strings.Contains(out, `/api/chat/stream`) {
		t.Fatalf("expected chat UI in output, got snippet: %.300s", out)
	}
	if strings.Contains(out, "{{.") {
		t.Fatalf("unexpanded template action in output")
	}
	if strings.Contains(out, `id="display-name"`) {
		t.Fatalf("display name input should not be rendered")
	}
}

func TestSendUploadServesMediaInline(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	sendUpload(rec, "photo.png", []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'})

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Disposition"); !strings.HasPrefix(got, "inline;") {
		t.Fatalf("content disposition = %q, want inline", got)
	}
	if got := resp.Header.Get("Content-Type"); got != "image/png" {
		t.Fatalf("content type = %q, want image/png", got)
	}
}

func TestSendUploadKeepsOtherFilesDownloadable(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	sendUpload(rec, "notes.txt", []byte("hello"))

	resp := rec.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := resp.Header.Get("Content-Disposition"); !strings.HasPrefix(got, "attachment;") {
		t.Fatalf("content disposition = %q, want attachment", got)
	}
	if got := resp.Header.Get("Content-Type"); got != "application/octet-stream" {
		t.Fatalf("content type = %q, want application/octet-stream", got)
	}
}

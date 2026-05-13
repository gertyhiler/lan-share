package httpadapter

import (
	"bytes"
	"html/template"
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
}

package httpadapter

import (
	"bytes"
	"encoding/base64"
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
	payload := `</textarea><script>bad</script>` + "\nкириллица 🎉"
	var buf bytes.Buffer
	err = tpl.Execute(&buf, indexData{
		ShareHost:      "192.168.1.10",
		Port:           8000,
		LatestPasteB64: base64.StdEncoding.EncodeToString([]byte(payload)),
		UploadsDir:     "/tmp/uploads",
		SharedDir:      "/tmp/shared",
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `data-initial="`) || !strings.Contains(out, `</textarea>`) {
		t.Fatalf("expected data-initial + textarea in output, got snippet: %.300s", out)
	}
	if strings.Contains(out, "{{.") {
		t.Fatalf("unexpanded template action in output")
	}
}

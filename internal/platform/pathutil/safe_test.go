package pathutil

import (
	"strings"
	"testing"
)

func TestSafeFilename(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want string
	}{
		{"", ""}, // non-empty after time suffix — actually empty becomes upload-timestamp; skip exact match
		{"../../etc/passwd", "passwd"},
		{"file  name.txt", "file name.txt"},
		{"weird:name?.txt", "weird_name_.txt"},
	}
	for _, tt := range tests {
		got := SafeFilename(tt.in)
		if tt.in == "" {
			if got == "" || !strings.HasPrefix(got, "upload-") {
				t.Fatalf("SafeFilename(%q) = %q, want upload-*", tt.in, got)
			}
			continue
		}
		if got != tt.want {
			t.Fatalf("SafeFilename(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

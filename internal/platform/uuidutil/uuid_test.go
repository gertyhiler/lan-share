package uuidutil

import "testing"

func TestNewV4ReturnsValidUUID(t *testing.T) {
	t.Parallel()
	id, err := NewV4()
	if err != nil {
		t.Fatalf("NewV4: %v", err)
	}
	if !IsValid(id) {
		t.Fatalf("invalid uuid: %q", id)
	}
}

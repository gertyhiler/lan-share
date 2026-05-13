package fs

import (
	"context"
	"testing"
	"time"

	"github.com/gertyhiler/lan-share/internal/domain"
)

func TestChatDeviceBindingsPersist(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := &Chat{Dir: t.TempDir()}

	if _, ok, err := store.DeviceIDForIP(ctx, "192.168.1.20"); err != nil || ok {
		t.Fatalf("initial binding = ok %v, err %v", ok, err)
	}
	if err := store.SaveDeviceIDForIP(ctx, "192.168.1.20", "00000000-0000-4000-8000-000000000001"); err != nil {
		t.Fatalf("save binding: %v", err)
	}
	got, ok, err := store.DeviceIDForIP(ctx, "192.168.1.20")
	if err != nil {
		t.Fatalf("read binding: %v", err)
	}
	if !ok || got != "00000000-0000-4000-8000-000000000001" {
		t.Fatalf("binding = %q, ok %v", got, ok)
	}
}

func TestChatRecentMessagesReturnsTail(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := &Chat{Dir: t.TempDir()}

	for _, id := range []string{"1", "2", "3"} {
		err := store.AppendMessage(ctx, domain.Message{
			ID:          id,
			TS:          time.Unix(1700000000, 0).UTC(),
			DeviceID:    "00000000-0000-4000-8000-000000000001",
			DisplayName: "Device",
			Text:        "msg " + id,
		})
		if err != nil {
			t.Fatalf("append %s: %v", id, err)
		}
	}

	got, err := store.RecentMessages(ctx, 2)
	if err != nil {
		t.Fatalf("recent messages: %v", err)
	}
	if len(got) != 2 || got[0].ID != "2" || got[1].ID != "3" {
		t.Fatalf("tail = %#v", got)
	}
}

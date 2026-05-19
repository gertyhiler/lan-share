package chat

import (
	"context"
	"testing"

	"github.com/gertyhiler/lan-share/internal/domain"
)

type memoryChatStore struct {
	messages []domain.Message
	devices  map[string]string
}

func (s *memoryChatStore) AppendMessage(_ context.Context, msg domain.Message) error {
	s.messages = append(s.messages, msg)
	return nil
}

func (s *memoryChatStore) RecentMessages(_ context.Context, limit int) ([]domain.Message, error) {
	if limit >= len(s.messages) {
		return s.messages, nil
	}
	return s.messages[len(s.messages)-limit:], nil
}

func (s *memoryChatStore) DeviceIDForIP(_ context.Context, ip string) (string, bool, error) {
	id, ok := s.devices[ip]
	return id, ok, nil
}

func (s *memoryChatStore) SaveDeviceIDForIP(_ context.Context, ip string, deviceID string) error {
	if s.devices == nil {
		s.devices = map[string]string{}
	}
	s.devices[ip] = deviceID
	return nil
}

func TestDeviceDisplayNameStable(t *testing.T) {
	t.Parallel()

	deviceID := "00000000-0000-4000-8000-000000000001"
	first := DeviceDisplayName(deviceID)
	second := DeviceDisplayName(deviceID)

	if first == "" {
		t.Fatal("device display name is empty")
	}
	if first != second {
		t.Fatalf("device display name is not stable: got %q and %q", first, second)
	}
}

func TestPostMessageUsesDeviceDisplayName(t *testing.T) {
	t.Parallel()

	store := &memoryChatStore{}
	service := NewService(store)
	deviceID := "00000000-0000-4000-8000-000000000001"

	msg, err := service.PostMessage(context.Background(), deviceID, "hello", nil)
	if err != nil {
		t.Fatalf("post message: %v", err)
	}

	want := DeviceDisplayName(deviceID)
	if msg.DisplayName != want {
		t.Fatalf("display name = %q, want %q", msg.DisplayName, want)
	}
}

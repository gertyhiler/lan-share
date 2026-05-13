package chat

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gertyhiler/lan-share/internal/domain"
	"github.com/gertyhiler/lan-share/internal/platform/pathutil"
	"github.com/gertyhiler/lan-share/internal/platform/uuidutil"
)

const (
	defaultDisplayName = "Устройство"
	maxDisplayNameLen  = 48
	maxMessageTextLen  = 16 << 10
)

// Service owns chat message creation and storage.
type Service struct {
	store domain.ChatStore
}

// NewService wires chat use cases.
func NewService(store domain.ChatStore) *Service {
	return &Service{store: store}
}

// DeviceIDForIP returns a stable device id previously assigned to a LAN IP.
func (s *Service) DeviceIDForIP(ctx context.Context, ip string) (string, bool, error) {
	return s.store.DeviceIDForIP(ctx, ip)
}

// SaveDeviceIDForIP persists a LAN IP to device id binding.
func (s *Service) SaveDeviceIDForIP(ctx context.Context, ip string, deviceID string) error {
	if !uuidutil.IsValid(deviceID) {
		return fmt.Errorf("invalid device id")
	}
	return s.store.SaveDeviceIDForIP(ctx, ip, deviceID)
}

// NewDeviceID creates a fresh server-owned device id.
func (s *Service) NewDeviceID() (string, error) {
	return uuidutil.NewV4()
}

// IsDeviceID reports whether s is a valid device id.
func IsDeviceID(s string) bool {
	return uuidutil.IsValid(s)
}

// RecentMessages returns the latest persisted chat messages.
func (s *Service) RecentMessages(ctx context.Context, limit int) ([]domain.Message, error) {
	return s.store.RecentMessages(ctx, limit)
}

// PostMessage validates, normalizes and appends a chat message.
func (s *Service) PostMessage(ctx context.Context, deviceID string, displayName string, text string, attachments []domain.Attachment) (domain.Message, error) {
	if !uuidutil.IsValid(deviceID) {
		return domain.Message{}, fmt.Errorf("invalid device id")
	}

	text = strings.TrimSpace(limitString(text, maxMessageTextLen))
	cleanAttachments := normalizeAttachments(attachments)
	if text == "" && len(cleanAttachments) == 0 {
		return domain.Message{}, fmt.Errorf("empty message")
	}

	id, err := uuidutil.NewV4()
	if err != nil {
		return domain.Message{}, err
	}
	msg := domain.Message{
		ID:          id,
		TS:          time.Now().UTC(),
		DeviceID:    deviceID,
		DisplayName: NormalizeDisplayName(displayName),
		Text:        text,
		Attachments: cleanAttachments,
	}
	if err := s.store.AppendMessage(ctx, msg); err != nil {
		return domain.Message{}, fmt.Errorf("append chat message: %w", err)
	}
	return msg, nil
}

// NormalizeDisplayName keeps UI-provided display names small and presentable.
func NormalizeDisplayName(name string) string {
	name = strings.TrimSpace(limitString(name, maxDisplayNameLen))
	if name == "" {
		return defaultDisplayName
	}
	return name
}

func normalizeAttachments(in []domain.Attachment) []domain.Attachment {
	out := make([]domain.Attachment, 0, len(in))
	for _, a := range in {
		name := pathutil.SafeFilename(a.Name)
		if name == "" {
			continue
		}
		out = append(out, domain.Attachment{
			Name:  name,
			Bytes: maxInt64(a.Bytes, 0),
			URL:   "/files/" + url.PathEscape(name),
		})
	}
	return out
}

func limitString(s string, maxRunes int) string {
	if maxRunes <= 0 || utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	n := 0
	for _, r := range s {
		if n >= maxRunes {
			break
		}
		b.WriteRune(r)
		n++
	}
	return b.String()
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

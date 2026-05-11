package paste

import (
	"context"
	"fmt"

	"github.com/gertyhiler/lan-share/internal/domain"
)

// Service saves and retrieves the latest text paste.
type Service struct {
	store domain.PasteStore
}

// NewService constructs a paste use case service.
func NewService(store domain.PasteStore) *Service {
	return &Service{store: store}
}

// Save persists a new paste (timestamped copy + latest).
func (s *Service) Save(ctx context.Context, text string) error {
	if err := s.store.Save(ctx, text); err != nil {
		return fmt.Errorf("paste save: %w", err)
	}
	return nil
}

// Latest returns the most recently saved paste text.
func (s *Service) Latest(ctx context.Context) (string, error) {
	text, err := s.store.Latest(ctx)
	if err != nil {
		return "", fmt.Errorf("paste latest: %w", err)
	}
	return text, nil
}

package files

import (
	"context"
	"fmt"

	"github.com/gertyhiler/lan-share/internal/domain"
)

// Service lists and transfers uploaded and shared files.
type Service struct {
	uploads domain.UploadStore
	shared  domain.SharedStore
}

// NewService wires file use cases.
func NewService(uploads domain.UploadStore, shared domain.SharedStore) *Service {
	return &Service{uploads: uploads, shared: shared}
}

// ListUploads lists files in the uploads directory.
func (s *Service) ListUploads(ctx context.Context) ([]domain.FileEntry, error) {
	list, err := s.uploads.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list uploads: %w", err)
	}
	return list, nil
}

// ListShared lists files in the shared directory (creates dir if missing).
func (s *Service) ListShared(ctx context.Context) ([]domain.FileEntry, error) {
	if err := s.shared.EnsureDir(ctx); err != nil {
		return nil, fmt.Errorf("shared ensure dir: %w", err)
	}
	list, err := s.shared.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list shared: %w", err)
	}
	return list, nil
}

// SaveUpload saves one uploaded file under a safe name.
func (s *Service) SaveUpload(ctx context.Context, filename string, data []byte) error {
	if err := s.uploads.SaveFile(ctx, filename, data); err != nil {
		return fmt.Errorf("save upload: %w", err)
	}
	return nil
}

// ReadUpload returns bytes for a file in the uploads directory.
func (s *Service) ReadUpload(ctx context.Context, name string) ([]byte, error) {
	b, err := s.uploads.ReadFile(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}
	return b, nil
}

// ReadShared returns bytes for a file in the shared directory.
func (s *Service) ReadShared(ctx context.Context, name string) ([]byte, error) {
	if err := s.shared.EnsureDir(ctx); err != nil {
		return nil, fmt.Errorf("shared ensure dir: %w", err)
	}
	b, err := s.shared.ReadFile(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("read shared: %w", err)
	}
	return b, nil
}

package domain

import "context"

// PasteStore persists shared text pastes.
type PasteStore interface {
	Save(ctx context.Context, text string) error
	Latest(ctx context.Context) (text string, err error)
}

// UploadStore stores files uploaded from the LAN into the uploads directory.
type UploadStore interface {
	List(ctx context.Context) ([]FileEntry, error)
	SaveFile(ctx context.Context, name string, data []byte) error
	ReadFile(ctx context.Context, name string) ([]byte, error)
}

// SharedStore reads files the operator placed in the shared directory for download.
type SharedStore interface {
	EnsureDir(ctx context.Context) error
	List(ctx context.Context) ([]FileEntry, error)
	ReadFile(ctx context.Context, name string) ([]byte, error)
}

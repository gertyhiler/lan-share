package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/gertyhiler/lan-share/internal/domain"
	"github.com/gertyhiler/lan-share/internal/platform/pathutil"
)

func listFiles(dir string) ([]domain.FileEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.FileEntry{}, nil
		}
		return nil, err
	}
	out := make([]domain.FileEntry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, domain.FileEntry{
			Name:  e.Name(),
			Bytes: info.Size(),
			MTime: info.ModTime().Unix(),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

// Pastes implements domain.PasteStore on disk.
type Pastes struct {
	Dir string
}

func (p *Pastes) Save(ctx context.Context, text string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	if err := os.MkdirAll(p.Dir, 0o755); err != nil {
		return fmt.Errorf("mkdir pastes: %w", err)
	}
	stamp := time.Now().Format("20060102-150405")
	path := filepath.Join(p.Dir, stamp+".txt")
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return fmt.Errorf("write paste: %w", err)
	}
	latest := filepath.Join(p.Dir, "latest.txt")
	if err := os.WriteFile(latest, []byte(text), 0o644); err != nil {
		return fmt.Errorf("write latest: %w", err)
	}
	return nil
}

func (p *Pastes) Latest(ctx context.Context) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}
	path := filepath.Join(p.Dir, "latest.txt")
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", domain.ErrNotFound
		}
		return "", err
	}
	return string(b), nil
}

// Uploads implements domain.UploadStore.
type Uploads struct {
	Dir string
}

func (u *Uploads) List(ctx context.Context) ([]domain.FileEntry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return listFiles(u.Dir)
}

func (u *Uploads) SaveFile(ctx context.Context, name string, data []byte) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	safe := pathutil.SafeFilename(name)
	if err := os.MkdirAll(u.Dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(u.Dir, safe)
	return os.WriteFile(path, data, 0o644)
}

func (u *Uploads) ReadFile(ctx context.Context, name string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	safe := pathutil.SafeFilename(name)
	path := filepath.Join(u.Dir, safe)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

// Shared implements domain.SharedStore.
type Shared struct {
	Dir string
}

func (s *Shared) EnsureDir(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return os.MkdirAll(s.Dir, 0o755)
}

func (s *Shared) List(ctx context.Context) ([]domain.FileEntry, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return listFiles(s.Dir)
}

func (s *Shared) ReadFile(ctx context.Context, name string) ([]byte, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	safe := pathutil.SafeFilename(name)
	path := filepath.Join(s.Dir, safe)
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return b, nil
}

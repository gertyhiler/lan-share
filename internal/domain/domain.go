package domain

import "errors"

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")

// FileEntry describes a file exposed via the HTTP API (mirrors Python JSON shape).
type FileEntry struct {
	Name  string `json:"name"`
	Bytes int64  `json:"bytes"`
	MTime int64  `json:"mtime"`
}

package fs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/gertyhiler/lan-share/internal/domain"
)

// Chat stores chat messages as JSONL and device bindings as JSON.
type Chat struct {
	Dir string
	mu  sync.Mutex
}

func (c *Chat) AppendMessage(ctx context.Context, msg domain.Message) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return fmt.Errorf("mkdir chat: %w", err)
	}
	f, err := os.OpenFile(c.messagesPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("open messages: %w", err)
	}
	defer func() { _ = f.Close() }()

	enc := json.NewEncoder(f)
	if err := enc.Encode(msg); err != nil {
		return fmt.Errorf("write message: %w", err)
	}
	return nil
}

func (c *Chat) RecentMessages(ctx context.Context, limit int) ([]domain.Message, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	if limit <= 0 {
		return []domain.Message{}, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	f, err := os.Open(c.messagesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return []domain.Message{}, nil
		}
		return nil, fmt.Errorf("open messages: %w", err)
	}
	defer func() { _ = f.Close() }()

	var out []domain.Message
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 2<<20)
	for sc.Scan() {
		var msg domain.Message
		if err := json.Unmarshal(sc.Bytes(), &msg); err != nil {
			continue
		}
		out = append(out, msg)
		if len(out) > limit {
			copy(out, out[len(out)-limit:])
			out = out[:limit]
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scan messages: %w", err)
	}
	return out, nil
}

func (c *Chat) DeviceIDForIP(ctx context.Context, ip string) (string, bool, error) {
	select {
	case <-ctx.Done():
		return "", false, ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	devices, err := c.readDevicesLocked()
	if err != nil {
		return "", false, err
	}
	id, ok := devices[ip]
	return id, ok, nil
}

func (c *Chat) SaveDeviceIDForIP(ctx context.Context, ip string, deviceID string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	devices, err := c.readDevicesLocked()
	if err != nil {
		return err
	}
	devices[ip] = deviceID
	return c.writeDevicesLocked(devices)
}

func (c *Chat) readDevicesLocked() (map[string]string, error) {
	b, err := os.ReadFile(c.devicesPath())
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("read devices: %w", err)
	}
	devices := map[string]string{}
	if len(bytes.TrimSpace(b)) == 0 {
		return devices, nil
	}
	if err := json.Unmarshal(b, &devices); err != nil {
		return nil, fmt.Errorf("parse devices: %w", err)
	}
	return devices, nil
}

func (c *Chat) writeDevicesLocked(devices map[string]string) error {
	if err := os.MkdirAll(c.Dir, 0o755); err != nil {
		return fmt.Errorf("mkdir chat: %w", err)
	}
	data, err := json.MarshalIndent(devices, "", "  ")
	if err != nil {
		return fmt.Errorf("encode devices: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(c.Dir, ".devices-*.json")
	if err != nil {
		return fmt.Errorf("create devices temp: %w", err)
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write devices temp: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close devices temp: %w", err)
	}
	if err := os.Rename(tmpName, c.devicesPath()); err != nil {
		return fmt.Errorf("rename devices: %w", err)
	}
	return nil
}

func (c *Chat) messagesPath() string {
	return filepath.Join(c.Dir, "messages.jsonl")
}

func (c *Chat) devicesPath() string {
	return filepath.Join(c.Dir, "devices.json")
}

package docs

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Cache stores and retrieves documentation by content hash.
type Cache struct {
	dir string
}

// NewCache creates a Cache that stores entries under dir.
func NewCache(dir string) *Cache {
	return &Cache{dir: dir}
}

// Get retrieves cached documentation for the given profile and source content.
func (c *Cache) Get(profileName string, content []byte) (string, bool) {
	key := c.key(profileName, content)
	path := filepath.Join(c.dir, key)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

// Set stores documentation for the given profile and source content.
func (c *Cache) Set(profileName string, content []byte, doc string) error {
	key := c.key(profileName, content)
	path := filepath.Join(c.dir, key)
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(doc), 0600); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	return nil
}

func (c *Cache) key(profileName string, content []byte) string {
	name := strings.TrimSpace(profileName)
	if name == "" {
		name = DefaultProfileName
	}
	h := sha256.New()
	_, _ = h.Write([]byte("goscribe-cache-v2\nprofile:"))
	_, _ = h.Write([]byte(name))
	_, _ = h.Write([]byte("\ncontent:"))
	_, _ = h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

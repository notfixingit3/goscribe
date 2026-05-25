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

// Get retrieves cached documentation for the given profile, template, and source content.
func (c *Cache) Get(profileName, templateName string, content []byte) (string, bool) {
	key := c.key(profileName, templateName, content)
	path := filepath.Join(c.dir, key)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(data), true
}

// Set stores documentation for the given profile, template, and source content.
func (c *Cache) Set(profileName, templateName string, content []byte, doc string) error {
	key := c.key(profileName, templateName, content)
	path := filepath.Join(c.dir, key)
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(doc), 0600); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	return nil
}

func (c *Cache) key(profileName, templateName string, content []byte) string {
	profile := strings.TrimSpace(profileName)
	if profile == "" {
		profile = DefaultProfileName
	}
	template := strings.TrimSpace(templateName)
	if template == "" {
		template = "none"
	}
	h := sha256.New()
	_, _ = h.Write([]byte("goscribe-cache-v3\nprofile:"))
	_, _ = h.Write([]byte(profile))
	_, _ = h.Write([]byte("\ntemplate:"))
	_, _ = h.Write([]byte(template))
	_, _ = h.Write([]byte("\ncontent:"))
	_, _ = h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

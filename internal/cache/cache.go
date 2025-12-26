// Package cache provides mechanisms for persisting and retrieving data locally
// to minimize network overhead and respect API rate limits.
package cache

import (
	"os"
	"path/filepath"
	"time"
)

// Cacher defines the behavior for data persistence and expiration logic.
type Cacher interface {
	// Get retrieves the cached data as a string.
	Get() ([]byte, error)
	// Set persists the provided string data to the cache.
	Set(data []byte) error
	// Expired returns true if the cached data is older than the specified TTL.
	Expired(ttl time.Duration) bool
}

// FileCacher is a filesystem-based implementation of the Cacher interface.
// It stores data in a specific file path on the local machine.
type FileCacher struct {
	// Path is the absolute path to the cache file.
	Path string
}

// NewFileCacher initializes a FileCacher. It determines the OS-specific user
// cache directory and appends the "nviman" namespace and the provided
// filename.
func NewFileCacher(dir string, filename string) *FileCacher {
	path := filepath.Join(dir, filename)
	return &FileCacher{Path: path}
}

// Get reads the file content from the disk and returns it as a byte slice.
// It returns an error if the file cannot be read.
func (fc *FileCacher) Get() ([]byte, error) {
	data, err := os.ReadFile(fc.Path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Set writes the provided byte slice to the filesystem.
// It automatically creates the necessary directory tree with 0755 permissions.
func (fc *FileCacher) Set(data []byte) error {
	if err := os.MkdirAll(filepath.Dir(fc.Path), 0755); err != nil {
		return err
	}
	return os.WriteFile(fc.Path, data, 0644)
}

// Expired checks the file modification time against the current time.
// It returns true if the duration since the last modification exceeds the TTL,
// or if the file does not exist.
func (fc *FileCacher) Expired(ttl time.Duration) bool {
	info, err := os.Stat(fc.Path)
	if err != nil {
		return true
	}
	return time.Since(info.ModTime()) > ttl
}

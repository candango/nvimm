package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileCacher(t *testing.T) {

	tmpDir := t.TempDir()
	filename := "test_release.json"
	c := NewFileCacher(tmpDir, filename)

	t.Run("should set and get data", func(t *testing.T) {
		expected := []byte(`[{"tag_name": "v0.11.3"}]`)

		err := c.Set(expected)
		if err != nil {
			t.Fatalf("failed to set cache: %v", err)
		}

		got, err := c.Get()
		if err != nil {
			t.Fatalf("failed to get cache: %v", err)
		}

		assert.Equal(t, expected, got)
	})
}

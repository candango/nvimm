package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/candango/nvimim/internal/config"
	"github.com/stretchr/testify/require"
)

func TestCurrentCommandIsReadOnly(t *testing.T) {
	installPath := t.TempDir()
	releasePath := filepath.Join(installPath, "0.12.0")
	require.NoError(t, os.Mkdir(releasePath, 0o755))
	require.NoError(t, os.Symlink(releasePath,
		filepath.Join(installPath, "current")))

	cmd := &CurrentCommand{}
	cmd.SetAppOptions(&config.AppOptions{Path: installPath})

	err := cmd.Execute([]string{"0.12.0"})
	require.EqualError(t, err,
		"current does not accept arguments; use 'set <release>'")

	current, err := os.Readlink(filepath.Join(installPath, "current"))
	require.NoError(t, err)
	require.Equal(t, releasePath, current)
}

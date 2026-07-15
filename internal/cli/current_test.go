package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/candango/nvimim/internal/config"
	"github.com/stretchr/testify/require"
)

const (
	explicitVersion    = "0.12.0"
	olderStableVersion = "0.11.5"
	prereleaseVersion  = "0.13.0-rc1"
)

func TestSetCommandSelectsInstalledRelease(t *testing.T) {
	installPath := t.TempDir()
	releasePath := filepath.Join(installPath, explicitVersion)
	otherReleasePath := filepath.Join(installPath, olderStableVersion)
	require.NoError(t, os.Mkdir(releasePath, 0o755))
	require.NoError(t, os.Mkdir(otherReleasePath, 0o755))

	cmd := &SetCommand{}
	cmd.SetAppOptions(&config.AppOptions{Path: installPath})

	require.NoError(t, cmd.Execute([]string{explicitVersion}))

	current, err := os.Readlink(filepath.Join(installPath, "current"))
	require.NoError(t, err)
	require.Equal(t, releasePath, current)
}

func TestSetCommandLatestSelectsHighestStableRelease(t *testing.T) {
	installPath := t.TempDir()
	for _, name := range []string{
		olderStableVersion,
		explicitVersion,
		"nightly",
		prereleaseVersion,
	} {
		require.NoError(t, os.Mkdir(filepath.Join(installPath, name), 0o755))
	}

	cmd := &SetCommand{}
	cmd.SetAppOptions(&config.AppOptions{Path: installPath})

	require.NoError(t, cmd.Execute([]string{"latest"}))

	current, err := os.Readlink(filepath.Join(installPath, "current"))
	require.NoError(t, err)
	require.Equal(t, filepath.Join(installPath, explicitVersion), current)
}

func TestSetCommandOnlySelectsInstalledReleases(t *testing.T) {
	installPath := t.TempDir()
	cmd := &SetCommand{}
	cmd.SetAppOptions(&config.AppOptions{Path: installPath})

	err := cmd.Execute([]string{explicitVersion})
	require.EqualError(t, err,
		"the release "+explicitVersion+" is not installed")
}

func TestCurrentCommandIsReadOnly(t *testing.T) {
	installPath := t.TempDir()
	releasePath := filepath.Join(installPath, explicitVersion)
	require.NoError(t, os.Mkdir(releasePath, 0o755))
	require.NoError(t, os.Symlink(releasePath,
		filepath.Join(installPath, "current")))

	cmd := &CurrentCommand{}
	cmd.SetAppOptions(&config.AppOptions{Path: installPath})

	err := cmd.Execute([]string{explicitVersion})
	require.EqualError(t, err,
		"current does not accept arguments; use 'set <release>'")

	current, err := os.Readlink(filepath.Join(installPath, "current"))
	require.NoError(t, err)
	require.Equal(t, releasePath, current)
}

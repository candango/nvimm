package cli

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallCommand_NoRelease(t *testing.T) {
	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	cmd := &InstallCommand{appOpts: opts}
	err := cmd.Execute(nil)
	assert.Error(t, err)
}

func TestInstallCommand_ReleaseNotFound(t *testing.T) {
	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	writeCache(t, opts.CachePath, releasesJSON("0.11.5", "nvim-linux-x86_64.tar.gz", "http://unused", "sha256:unused"))

	cmd := &InstallCommand{appOpts: opts}
	err := cmd.Execute([]string{"9.9.9"})
	assert.Error(t, err)
}

func TestInstallCommand_DeclineDownload(t *testing.T) {
	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	version := "0.11.5"
	writeCache(t, opts.CachePath, releasesJSON(version, "nvim-linux-x86_64.tar.gz", "http://unused", "sha256:unused"))

	cmd := &InstallCommand{
		appOpts: opts,
		stdin:   strings.NewReader("n\n"),
	}
	err := cmd.Execute([]string{version})
	assert.NoError(t, err)
	assert.NoDirExists(t, filepath.Join(opts.Path, version))
}

func TestInstallCommand_YesFlag(t *testing.T) {
	name := assetName()
	if name == "" {
		t.Skipf("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	version := "0.11.5"
	tarPath, digest := makeFakeTarball(t, t.TempDir(), name, version)
	tarData, err := os.ReadFile(tarPath)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, name) {
			w.Write(tarData)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	writeCache(t, opts.CachePath, releasesJSON(version, name, srv.URL, digest))

	cmd := &InstallCommand{appOpts: opts, Yes: true}
	err = cmd.Execute([]string{version})
	assert.NoError(t, err)
	assert.DirExists(t, filepath.Join(opts.Path, version))
	assert.Equal(t, version, readCurrentVersion(t, opts.Path))
}

func TestInstallCommand_DeclineSetCurrent(t *testing.T) {
	name := assetName()
	if name == "" {
		t.Skipf("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	version := "0.11.5"
	tarPath, digest := makeFakeTarball(t, t.TempDir(), name, version)
	tarData, err := os.ReadFile(tarPath)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, name) {
			w.Write(tarData)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	writeCache(t, opts.CachePath, releasesJSON(version, name, srv.URL, digest))

	// y to download, n to set current
	cmd := &InstallCommand{
		appOpts: opts,
		stdin:   strings.NewReader("y\nn\n"),
	}
	err = cmd.Execute([]string{version})
	assert.NoError(t, err)
	assert.DirExists(t, filepath.Join(opts.Path, version))
	_, linkErr := os.Readlink(filepath.Join(opts.Path, "current"))
	assert.True(t, os.IsNotExist(linkErr), "current symlink should not exist")
}

func TestInstallCommand_StableAlias(t *testing.T) {
	name := assetName()
	if name == "" {
		t.Skipf("unsupported platform %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	version := "0.11.5"
	tarPath, digest := makeFakeTarball(t, t.TempDir(), name, version)
	tarData, err := os.ReadFile(tarPath)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, name) {
			w.Write(tarData)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	writeCache(t, opts.CachePath, releasesJSON(version, name, srv.URL, digest))

	// "stable" resolves to 0.11.5 via Releases.Get("stable")
	cmd := &InstallCommand{appOpts: opts, Yes: true}
	err = cmd.Execute([]string{"stable"})
	assert.NoError(t, err)
	assert.DirExists(t, filepath.Join(opts.Path, version))
}

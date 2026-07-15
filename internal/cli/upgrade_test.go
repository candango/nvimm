package cli

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/candango/nvimim/internal/cache"
	"github.com/candango/nvimim/internal/config"
	"github.com/candango/nvimim/internal/release"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// releasesJSON builds a minimal GitHub releases JSON for tests.
// stableVersion is the clean version string (e.g. "0.11.5").
// assetName is the expected tarball filename.
// htmlURL is the base HTML URL (e.g. from an httptest server).
func releasesJSON(stableVersion, assetName, htmlURL, digest string) string {
	tag := "v" + stableVersion
	return fmt.Sprintf(`[
		{
			"tag_name": "stable",
			"name": "NVIM v%s",
			"html_url": "%s/releases/tag/stable"
		},
		{
			"tag_name": "%s",
			"name": "NVIM v%s",
			"html_url": "%s/releases/tag/%s",
			"assets": [
				{
					"id": 1,
					"name": "%s",
					"digest": "%s",
					"state": "uploaded",
					"content_type": "application/gzip"
				}
			]
		}
	]`, stableVersion, htmlURL, tag, stableVersion, htmlURL, tag, assetName, digest)
}

// makeFakeTarball creates a minimal .tar.gz in destDir and returns its path
// and sha256 digest in "sha256:hex" format.
func makeFakeTarball(t *testing.T, destDir, filename, version string) (string, string) {
	t.Helper()

	tarPath := filepath.Join(destDir, filename)
	f, err := os.Create(tarPath)
	require.NoError(t, err)
	defer f.Close()

	h := sha256.New()
	mw := io.MultiWriter(f, h)

	gzw := gzip.NewWriter(mw)
	tw := tar.NewWriter(gzw)

	// dir name inside tarball must match filename without .tar.gz
	// so that releasePath in installRelease resolves correctly
	dirName := strings.TrimSuffix(filename, ".tar.gz")
	content := []byte("fake nvim binary")
	hdr := &tar.Header{
		Name: fmt.Sprintf("%s/bin/nvim", dirName),
		Mode: 0755,
		Size: int64(len(content)),
	}
	require.NoError(t, tw.WriteHeader(hdr))
	_, err = tw.Write(content)
	require.NoError(t, err)
	require.NoError(t, tw.Close())
	require.NoError(t, gzw.Close())
	require.NoError(t, f.Close())

	// recompute hash from file (MultiWriter hash is before gzip flush)
	fh, err := os.Open(tarPath)
	require.NoError(t, err)
	defer fh.Close()
	h2 := sha256.New()
	_, err = io.Copy(h2, fh)
	require.NoError(t, err)
	digest := fmt.Sprintf("sha256:%x", h2.Sum(nil))

	return tarPath, digest
}

// setupAppOpts creates temp dirs and returns a minimal AppOptions.
func setupAppOpts(t *testing.T) *config.AppOptions {
	t.Helper()
	tmp := t.TempDir()
	return &config.AppOptions{
		Path:       filepath.Join(tmp, "nvim"),
		CachePath:  filepath.Join(tmp, "cache"),
		MinRelease: "0.7.0",
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(path, 0755))
}

func writeCache(t *testing.T, cachePath, data string) {
	t.Helper()
	mustMkdir(t, cachePath)
	c := cache.NewFileCacher(cachePath, "nvimim_releases.json")
	require.NoError(t, c.Set([]byte(data)))
}

// currentVersion returns the clean tag for the symlink target.
func readCurrentVersion(t *testing.T, nvimPath string) string {
	t.Helper()
	link, err := os.Readlink(filepath.Join(nvimPath, "current"))
	require.NoError(t, err)
	return filepath.Base(link)
}

// --- prompt ---

func TestPrompt(t *testing.T) {
	cases := []struct {
		input    string
		expected bool
	}{
		{"y\n", true},
		{"yes\n", true},
		{"Y\n", true},
		{"YES\n", true},
		{"n\n", false},
		{"no\n", false},
		{"N\n", false},
		{"\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tc.input))
			got, err := prompt("test? ", r)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

// --- UpgradeCommand ---

func TestUpgradeCommand_AlreadyUpToDate(t *testing.T) {
	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	version := "0.11.5"
	mustMkdir(t, filepath.Join(opts.Path, version))
	require.NoError(t, os.Symlink(
		filepath.Join(opts.Path, version),
		filepath.Join(opts.Path, "current"),
	))

	writeCache(t, opts.CachePath, releasesJSON(version, "nvim-linux-x86_64.tar.gz", "http://unused", "sha256:unused"))

	cmd := &UpgradeCommand{appOpts: opts}
	err := cmd.Execute(nil)
	assert.NoError(t, err)
}

func TestUpgradeCommand_InstalledNotCurrent_AcceptSetCurrent(t *testing.T) {
	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	version := "0.11.5"
	mustMkdir(t, filepath.Join(opts.Path, version))
	// no "current" symlink

	writeCache(t, opts.CachePath, releasesJSON(version, "nvim-linux-x86_64.tar.gz", "http://unused", "sha256:unused"))

	cmd := &UpgradeCommand{
		appOpts: opts,
		stdin:   strings.NewReader("y\n"),
	}
	err := cmd.Execute(nil)
	assert.NoError(t, err)
	assert.Equal(t, version, readCurrentVersion(t, opts.Path))
}

func TestUpgradeCommand_InstalledNotCurrent_DeclineSetCurrent(t *testing.T) {
	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	version := "0.11.5"
	mustMkdir(t, filepath.Join(opts.Path, version))

	writeCache(t, opts.CachePath, releasesJSON(version, "nvim-linux-x86_64.tar.gz", "http://unused", "sha256:unused"))

	cmd := &UpgradeCommand{
		appOpts: opts,
		stdin:   strings.NewReader("n\n"),
	}
	err := cmd.Execute(nil)
	assert.NoError(t, err)
	_, err = os.Readlink(filepath.Join(opts.Path, "current"))
	assert.True(t, os.IsNotExist(err), "current symlink should not exist")
}

func TestUpgradeCommand_InstalledNotCurrent_YesFlag(t *testing.T) {
	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	version := "0.11.5"
	mustMkdir(t, filepath.Join(opts.Path, version))

	writeCache(t, opts.CachePath, releasesJSON(version, "nvim-linux-x86_64.tar.gz", "http://unused", "sha256:unused"))

	cmd := &UpgradeCommand{appOpts: opts, Yes: true}
	err := cmd.Execute(nil)
	assert.NoError(t, err)
	assert.Equal(t, version, readCurrentVersion(t, opts.Path))
}

func TestUpgradeCommand_NotInstalled_DeclineDownload(t *testing.T) {
	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	version := "0.11.5"
	// no version dir — not installed

	writeCache(t, opts.CachePath, releasesJSON(version, "nvim-linux-x86_64.tar.gz", "http://unused", "sha256:unused"))

	cmd := &UpgradeCommand{
		appOpts: opts,
		stdin:   strings.NewReader("n\n"),
	}
	err := cmd.Execute(nil)
	assert.NoError(t, err)
	assert.NoDirExists(t, filepath.Join(opts.Path, version))
}

// --- installRelease with HTTP mock ---

func assetName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH
	if goos == "linux" && goarch == "amd64" {
		return "nvim-linux-x86_64.tar.gz"
	}
	if goos == "linux" && goarch == "arm64" {
		return "nvim-linux-arm64.tar.gz"
	}
	if goos == "darwin" && goarch == "amd64" {
		return "nvim-macos-x86_64.tar.gz"
	}
	if goos == "darwin" && goarch == "arm64" {
		return "nvim-macos-arm64.tar.gz"
	}
	return ""
}

func TestInstallRelease_YesFlag(t *testing.T) {
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
			w.Header().Set("Content-Type", "application/gzip")
			w.WriteHeader(http.StatusOK)
			w.Write(tarData)
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	info := &release.Info{
		TagName: "v" + version,
		Name:    "NVIM v" + version,
		HtmlUrl: srv.URL + "/releases/tag/v" + version,
		Assets: []release.Asset{
			{Name: name, Digest: digest},
		},
	}

	err = installRelease(info, opts, nil, true)
	assert.NoError(t, err)
	assert.DirExists(t, filepath.Join(opts.Path, version))
	assert.Equal(t, version, readCurrentVersion(t, opts.Path))
}

func TestInstallRelease_DeclineDownload(t *testing.T) {
	opts := setupAppOpts(t)
	mustMkdir(t, opts.Path)
	mustMkdir(t, opts.CachePath)

	version := "0.11.5"
	info := &release.Info{
		TagName: "v" + version,
		Name:    "NVIM v" + version,
	}

	err := installRelease(info, opts, strings.NewReader("n\n"), false)
	assert.NoError(t, err)
	assert.NoDirExists(t, filepath.Join(opts.Path, version))
}

func TestInstallRelease_DeclineSetCurrent(t *testing.T) {
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

	info := &release.Info{
		TagName: "v" + version,
		Name:    "NVIM v" + version,
		HtmlUrl: srv.URL + "/releases/tag/v" + version,
		Assets:  []release.Asset{{Name: name, Digest: digest}},
	}

	// y to download, n to set current
	err = installRelease(info, opts, strings.NewReader("y\nn\n"), false)
	assert.NoError(t, err)
	assert.DirExists(t, filepath.Join(opts.Path, version))
	_, linkErr := os.Readlink(filepath.Join(opts.Path, "current"))
	assert.True(t, os.IsNotExist(linkErr), "current symlink should not exist")
}

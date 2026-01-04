package cli

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/candango/iook/dir"
	"github.com/candango/iook/pathx"
	"github.com/candango/nvimm/internal/cache"
	"github.com/candango/nvimm/internal/protocol"
	"github.com/candango/nvimm/internal/release"
)

type CurrentCommand struct {
}

func (cmd *CurrentCommand) Execute(args []string) error {
	return nil
}

type InstallCommand struct {
	Release    string `positional-arg-name:"release" description:"Release version to install"`
	appOptions *AppOptions
}

func (cmd *InstallCommand) Usage() string {
	return "<release>"
}

func (cmd *InstallCommand) Execute(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("positional argument release was not informed\n")
	}
	cmd.Release = args[0]
	if !pathx.Exists(cmd.appOptions.CachePath()) {
		return fmt.Errorf("cache path does not exist: %s",
			cmd.appOptions.CachePath())
	}
	cachePath := cmd.appOptions.CachePath()

	releaseCacher := cache.NewFileCacher(cachePath, "nvimm_releases.json")
	gt, err := protocol.NewGithubTransport()
	if err != nil {
		return fmt.Errorf("failed to create github transport: %w", err)
	}

	// TODO: use parametrized expiration time
	if releaseCacher.Expired(30 * time.Minute) {
		res, err := gt.GetReleases()
		if err != nil {
			return fmt.Errorf("failed to get releases: %w", err)
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		err = releaseCacher.Set(data)
		if err != nil {
			return fmt.Errorf("failed to cache releases: %w", err)
		}
	}

	data, err := releaseCacher.Get()
	if err != nil {
		return fmt.Errorf("failed to get cached releases: %w", err)
	}
	releases := release.Releases{}
	err = releases.Process(data)
	if err != nil {
		return fmt.Errorf("failed to process releases: %w", err)
	}
	info, err := releases.Get(cmd.Release)
	if err != nil {
		return err
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	assetUrl := ""
	assetFound := false

	for _, asset := range info.Assets {

		if asset.Name == getTarballName(goos, goarch) {
			assetFound = true
			assetUrl = fmt.Sprintf("%s/%s", strings.ReplaceAll(info.HtmlUrl, "tag", "download"), asset.Name)
		}
	}

	if !assetFound {
		return fmt.Errorf("the os %s and arch %s cannot to be resolved as a valid nvim asset", goos, goarch)
	}

	downloadedRelease, err := downloadRelease(assetUrl, cachePath)
	if err != nil {
		return err
	}

	fmt.Printf("the release %s was downloaded to %s/n", downloadedRelease, cachePath)

	untarFile(filepath.Join(cachePath, downloadedRelease))

	releasePath := strings.ReplaceAll(
		filepath.Join(cachePath, downloadedRelease), ".tar.gz", "")
	dir.CopyAll(releasePath, filepath.Join(cmd.appOptions.Path, cmd.Release))
	return nil
}

func getTarballName(goos string, goarch string) string {

	if goos == "darwin" && goarch == "amd64" {
		return "nvim-macos-x86_64.tar.gz"
	}

	if goos == "darwin" && goarch == "arm64" {
		return "nvim-macos-arm64.tar.gz"
	}

	if goos == "linux" && goarch == "amd64" {
		return "nvim-linux-x86_64.tar.gz"
	}

	if goos == "linux" && goarch == "arm64" {
		return "nvim-linux-arm64.tar.gz"
	}

	return ""
}

func untarFile(path string) error {
	destDir := filepath.Dir(path)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Sanitize path to prevent path traversal (security check)
		name := filepath.Clean(header.Name)
		if strings.Contains(name, "..") || filepath.IsAbs(name) {
			return fmt.Errorf(
				"security error: tar entry %q contains an invalid or unsafe "+
					"path (possible path traversal attempt), extraction "+
					"aborted", name)
		}
		target := filepath.Join(destDir, name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
			if err := os.Chmod(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
			// Set timestamps
			if err := os.Chtimes(target, header.AccessTime, header.ModTime); err != nil {
				// Not fatal, continue
			}
			// TODO: Set file ownership (UID/GID) if needed and running as root
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}
		case tar.TypeLink:
			linkTarget := filepath.Join(destDir, header.Linkname)
			if err := os.Link(linkTarget, target); err != nil {
				return err
			}
		}
	}
	return nil
}

func downloadRelease(url string, destDir string) (string, error) {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	filename := path.Base(url)
	outPath := path.Join(destDir, filename)
	out, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	if _, err = io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return filename, nil
}

func (cmd *InstallCommand) SetAppOptions(opts *AppOptions) {
	cmd.appOptions = opts
}

type ListCommand struct {
	appOptions *AppOptions
}

func (cmd *ListCommand) Execute(args []string) error {
	if !pathx.Exists(cmd.appOptions.CachePath()) {
		return fmt.Errorf("cache path does not exist: %s",
			cmd.appOptions.CachePath())
	}
	cachePath := cmd.appOptions.CachePath()
	releaseCacher := cache.NewFileCacher(cachePath, "nvimm_releases.json")
	gt, err := protocol.NewGithubTransport()
	if err != nil {
		return fmt.Errorf("failed to create github transport: %w", err)
	}

	// TODO: use parametrized expiration time
	if releaseCacher.Expired(30 * time.Minute) {
		res, err := gt.GetReleases()
		if err != nil {
			return fmt.Errorf("failed to get releases: %w", err)
		}
		data, err := io.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		err = releaseCacher.Set(data)
		if err != nil {
			return fmt.Errorf("failed to cache releases: %w", err)
		}
	}

	data, err := releaseCacher.Get()
	if err != nil {
		return fmt.Errorf("failed to get cached releases: %w", err)
	}

	releases := release.Releases{}
	err = releases.Process(data)
	if err != nil {
		return fmt.Errorf("failed to process releases: %w", err)
	}

	for _, info := range releases {
		if info.Stable == true {
			fmt.Printf("%s (stable)\n", info.CleanTagName())
			continue
		}
		fmt.Printf("%s\n", info.CleanTagName())
	}
	return nil
}

func (cmd *ListCommand) SetAppOptions(opts *AppOptions) {
	cmd.appOptions = opts
}

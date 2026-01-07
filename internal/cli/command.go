package cli

import (
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

	"github.com/candango/iook/archive"
	"github.com/candango/iook/dir"
	"github.com/candango/iook/pathx"
	"github.com/candango/nvimm/internal/cache"
	"github.com/candango/nvimm/internal/config"
	"github.com/candango/nvimm/internal/filehash"
	"github.com/candango/nvimm/internal/protocol"
	"github.com/candango/nvimm/internal/release"
)

type CurrentCommand struct {
	Release string `positional-arg-name:"release" description:"Release version to be set"`
	appOpts *config.AppOptions
}

func (cmd *CurrentCommand) Usage() string {
	return "[release]"
}

func (cmd *CurrentCommand) Execute(args []string) error {
	if !pathx.Exists(cmd.appOpts.CachePath) {
		return fmt.Errorf("cache path does not exist: %s",
			cmd.appOpts.CachePath)
	}
	if !pathx.Exists(cmd.appOpts.Path) {
		return fmt.Errorf("nvim path does not exist: %s",
			cmd.appOpts.Path)
	}
	// nvimPath := cmd.appOptions.Path
	cachePath := cmd.appOpts.CachePath

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

	err = releases.Process(data, cmd.appOpts)
	if err != nil {
		return fmt.Errorf("failed to process releases: %w", err)
	}
	notInstalled := len(releases.Installed(cmd.appOpts.Path)) == 0
	if notInstalled {
		return fmt.Errorf("no releases installed yet")
	}
	mustSetCurrent := true
	if len(args) == 0 {
		mustSetCurrent = false
	} else {
		cmd.Release = args[0]
		if !pathx.Exists(filepath.Join(cmd.appOpts.Path, cmd.Release)) {
			return fmt.Errorf("the release %s is not installed", cmd.Release)
		}

	}
	if !mustSetCurrent {
		currentInstalled, err := os.Readlink(filepath.Join(cmd.appOpts.Path, "current"))
		if err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to read current symlink: %w", err)
			}
			fmt.Printf("no current version set\n")
			return nil

		}
		fmt.Printf("* %s\n", filepath.Base(currentInstalled))
		return nil
	}

	currentInstalled, err := os.Readlink(filepath.Join(cmd.appOpts.Path, "current"))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read current symlink: %w", err)
		}
	}

	if currentInstalled == filepath.Join(cmd.appOpts.Path, cmd.Release) {
		fmt.Printf("the release %s is already set as current\n", cmd.Release)
		return nil
	}

	os.RemoveAll(filepath.Join(cmd.appOpts.Path, "current"))
	os.Symlink(
		filepath.Join(cmd.appOpts.Path, cmd.Release),
		filepath.Join(cmd.appOpts.Path, "current"))
	return nil
}

func (cmd *CurrentCommand) SetAppOptions(opts *config.AppOptions) {
	cmd.appOpts = opts
}

type InstallCommand struct {
	Release string `positional-arg-name:"release" description:"Release version to install"`
	appOpts *config.AppOptions
}

func (cmd *InstallCommand) Usage() string {
	return "<release>"
}

func (cmd *InstallCommand) Execute(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("positional argument release was not informed\n")
	}
	cmd.Release = args[0]
	if !pathx.Exists(cmd.appOpts.CachePath) {
		return fmt.Errorf("cache path does not exist: %s",
			cmd.appOpts.CachePath)
	}
	if !pathx.Exists(cmd.appOpts.Path) {
		return fmt.Errorf("nvim path does not exist: %s",
			cmd.appOpts.Path)
	}
	// nvimPath := cmd.appOptions.Path
	cachePath := cmd.appOpts.CachePath

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

	err = releases.Process(data, cmd.appOpts)
	if err != nil {
		return fmt.Errorf("failed to process releases: %w", err)
	}

	mustSetCurrent := len(releases.Installed(cmd.appOpts.Path)) == 0
	info, err := releases.Get(cmd.Release)
	if err != nil {
		return err
	}

	goos := runtime.GOOS
	goarch := runtime.GOARCH
	assetUrl := ""
	assetDigest := ""
	assetFound := false

	for _, asset := range info.Assets {
		if asset.Name == getTarballName(info, goos, goarch) {
			assetFound = true
			assetUrl = fmt.Sprintf("%s/%s", strings.ReplaceAll(info.HtmlUrl, "tag", "download"), asset.Name)
			assetDigest = asset.Digest
			break
		}
	}

	if !assetFound {
		return fmt.Errorf("the os %s and arch %s cannot to be resolved as a valid nvim asset", goos, goarch)
	}

	downloadedRelease, err := downloadRelease(assetUrl, cachePath)
	if err != nil {
		return err
	}
	downloadedFile := filepath.Join(cachePath, downloadedRelease)
	fmt.Printf("the release %s was downloaded to %s\n", downloadedRelease, cachePath)
	fingerprint, err := filehash.SHA256(downloadedFile)
	if err != nil {
		return err
	}

	if fingerprint != assetDigest {
		return fmt.Errorf("the downloaded file is corrupted: expected %s but got %s",
			assetDigest, fingerprint)
	}

	f, err := os.Open(downloadedFile)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()
	fmt.Printf("extracting to %s\n", filepath.Dir(downloadedFile))
	archive.Untar(gzr, filepath.Dir(downloadedFile))

	releasePath := strings.ReplaceAll(
		filepath.Join(cachePath, downloadedRelease), ".tar.gz", "")
	dir.CopyAll(releasePath, filepath.Join(cmd.appOpts.Path, cmd.Release))
	if mustSetCurrent {
		os.RemoveAll(filepath.Join(cmd.appOpts.Path, "current"))
		os.Symlink(
			filepath.Join(cmd.appOpts.Path, cmd.Release),
			filepath.Join(cmd.appOpts.Path, "current"))
	}

	return nil
}

func getTarballName(info *release.Info, goos string, goarch string) string {

	if goos == "darwin" && goarch == "amd64" {
		return "nvim-macos-x86_64.tar.gz"
	}

	if goos == "darwin" && goarch == "arm64" {
		return "nvim-macos-arm64.tar.gz"
	}

	if goos == "linux" && goarch == "amd64" {
		if info.VersionLess("0.10.4") {
			return "nvim-linux64.tar.gz"
		}
		return "nvim-linux-x86_64.tar.gz"
	}

	if goos == "linux" && goarch == "arm64" {
		return "nvim-linux-arm64.tar.gz"
	}

	return ""
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

func (cmd *InstallCommand) SetAppOptions(opts *config.AppOptions) {
	cmd.appOpts = opts
}

type ListCommand struct {
	appOpts *config.AppOptions
}

func (cmd *ListCommand) Execute(args []string) error {
	if !pathx.Exists(cmd.appOpts.CachePath) {
		return fmt.Errorf("cache path does not exist: %s",
			cmd.appOpts.CachePath)
	}
	cachePath := cmd.appOpts.CachePath
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
	err = releases.Process(data, cmd.appOpts)
	if err != nil {
		return fmt.Errorf("failed to process releases: %w", err)
	}

	installed := releases.Installed(cmd.appOpts.Path)

	fmt.Println("Installed versions")
	if len(installed) == 0 {
		fmt.Println("  no releases installed")
	}

	currentInstalled, err := os.Readlink(filepath.Join(cmd.appOpts.Path, "current"))
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read current symlink: %w", err)
		}
	}

	for _, info := range installed {
		ident := "  "
		if filepath.Base(currentInstalled) == info.CleanTagName() {
			ident = "* "
		}
		if info.Stable == true {
			fmt.Printf("%s%s (stable)\n", ident, info.CleanTagName())
			continue
		}
		fmt.Printf("%s%s\n", ident, info.CleanTagName())
	}

	available := releases.Available(installed)
	if len(available) > 0 {
		fmt.Println("\nAvailable versions")
	}
	for _, info := range available {
		if info.Stable == true {
			fmt.Printf("  %s (stable)\n", info.CleanTagName())
			continue
		}
		fmt.Printf("  %s\n", info.CleanTagName())
	}
	return nil
}

func (cmd *ListCommand) SetAppOptions(opts *config.AppOptions) {
	cmd.appOpts = opts
}

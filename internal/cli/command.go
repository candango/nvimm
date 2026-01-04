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
}

func (cmd *InstallCommand) Execute(args []string) error {
	return nil
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

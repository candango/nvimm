package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/candango/iook/pathx"
	"github.com/jessevdk/go-flags"
)

type AppOptions struct {
	Verbose        bool   `short:"v" long:"verbose" description:"Enable verbose mode"`
	CachePath      string `short:"C" long:"cache-path" env:"NVIMM_CACHE_PATH" description:"Cache directory"`
	ConfigPath     string `short:"c" long:"config" env:"NVIMM_CONFIG_PATH" description:"Configuration file path"`
	ConfigDir      string `short:"d" long:"config-dir" env:"NVIMM_CONFIG_DIR" description:"Configuration file directory"`
	ConfigFileName string `short:"n" long:"config-file-name" env:"NVIMM_CONFIG_FILE_NAME" default:"nvimm.yml" description:"Configuration file name"`
	Path           string `short:"p" long:"path" env:"NVIMM_PATH" description:"Path where Neovim releases are installed"`
	MinRelease     string `short:"r" long:"min-release" env:"NVIMM_MIN_RELEASE" default:"0.7.0" description:"Neovim minimal release"`
}

type AppOptionsAware interface {
	SetAppOptions(opts *AppOptions)
}

func WithError(err error) func(cmd flags.Commander, args []string) error {
	return func(_ flags.Commander, _ []string) error {
		return err
	}
}

type AppOptionsFunc func(opts *AppOptions) error

func WithAppOptions(opts *AppOptions, fns ...AppOptionsFunc) func(cmd flags.Commander, args []string) error {
	return func(cmd flags.Commander, args []string) error {
		if opts.ConfigDir == "" {
			userConfigDir, err := os.UserConfigDir()
			if err != nil {
				return err
			}
			opts.ConfigDir = filepath.Join(userConfigDir, "nvimm")
		}
		opts.ConfigPath = filepath.Join(opts.ConfigDir, opts.ConfigFileName)

		if opts.Path == "" {
			userHomeDir, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			opts.Path = filepath.Join(userHomeDir, ".nvimm")
		}

		if opts.CachePath == "" {
			userCacheDir, err := os.UserCacheDir()
			if err != nil {
				return err
			}
			opts.CachePath = filepath.Join(userCacheDir, "nvimm")
		}

		// Apply extra functions
		if len(fns) > 0 {
			for _, fn := range fns {
				err := fn(opts)
				if err != nil {
					return err
				}
			}
		}

		if aware, ok := cmd.(AppOptionsAware); ok == true {
			aware.SetAppOptions(opts)
		}
		return cmd.Execute(args)
	}
}

func WithPathsResolved(opts *AppOptions) error {
	if !pathx.Exists(opts.ConfigDir) {
		err := os.MkdirAll(opts.ConfigDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating nvimm config dir %s: %v",
				opts.ConfigDir, err)
		}
	}

	if !pathx.Exists(opts.ConfigPath) {
		_, err := os.Create(opts.ConfigPath)
		if err != nil {
			return fmt.Errorf("error creating nvimm config path %s: %v",
				opts.ConfigPath, err)
		}
		err = os.Chmod(opts.ConfigPath, 0644)
		if err != nil {
			return fmt.Errorf("error changing nvimm config path %s "+
				"permission:%v", opts.ConfigPath, err)
		}
	}

	if !pathx.Exists(opts.Path) {
		err := os.MkdirAll(opts.Path, 0755)
		if err != nil {
			return fmt.Errorf("error creating nvimm path %s: %v",
				opts.Path, err)
		}
	}

	if !pathx.Exists(opts.CachePath) {
		err := os.MkdirAll(opts.CachePath, 0755)
		if err != nil {
			return fmt.Errorf("error creating nvimm cache path %s: %v",
				opts.CachePath, err)
		}
	}
	return nil
}

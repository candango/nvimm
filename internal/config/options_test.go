package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
)

type OptionCheck struct {
	configDir string
}

type TestOptionsCommand struct {
	*AppOptions
}

func (cmd *TestOptionsCommand) SetAppOptions(opts *AppOptions) {
	cmd.AppOptions = opts
}

func (cmd *TestOptionsCommand) Execute(args []string) error {
	return nil
}

func TestOptions(t *testing.T) {

	t.Run("should get values from environment", func(t *testing.T) {
		var opts AppOptions
		os.Setenv("NVIMIM_CONFIG_DIR", "/etc/nvimim")
		os.Setenv("NVIMIM_CONFIG_FILE_NAME", "nvimim.yaml")
		os.Setenv("NVIMIM_PATH", "/opt/nvimim")
		os.Setenv("NVIMIM_CACHE_PATH", "/opt/nvim/cache")
		defer os.Unsetenv("NVIMIM_CONFIG_DIR")
		defer os.Unsetenv("NVIMIM_CONFIG_FILE_NAME")
		defer os.Unsetenv("NVIMIM_PATH")
		defer os.Unsetenv("NVIMIM_CACHE_PATH")

		parser := flags.NewParser(&opts, flags.Default)
		parser.Usage = "[Application|Help Options] command"

		handler := WithAppOptions(&opts)

		parser.CommandHandler = handler

		cmd := &TestOptionsCommand{}

		parser.AddCommand(
			"options",
			"Run test command to check options",
			"Run test command to check options",
			cmd)

		_, err := parser.ParseArgs([]string{"options"})
		if err != nil {
			t.Fatalf("error running the command: %v", err)
		}

		assert.Equal(t, filepath.Join(os.Getenv("NVIMIM_CONFIG_DIR"),
			os.Getenv("NVIMIM_CONFIG_FILE_NAME")), opts.ConfigPath)
		assert.Equal(t, os.Getenv("NVIMIM_PATH"), opts.Path)
		assert.Equal(t, opts.CachePath, opts.CachePath)
	})

	t.Run("should get default values", func(t *testing.T) {
		var opts AppOptions

		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			t.Fatalf("error getting user config dir: %v", err)
		}
		expectedConfigPath := filepath.Join(userConfigDir, "nvimim", "nvimim.yml")
		parser := flags.NewParser(&opts, flags.Default)

		parser.Usage = "[Application|Help Options] command"

		parser.CommandHandler = WithAppOptions(&opts)

		cmd := &TestOptionsCommand{}

		parser.AddCommand(
			"options",
			"Run test command to check options",
			"Run test command to check options",
			cmd)

		_, err = parser.ParseArgs([]string{"options"})
		if err != nil {
			t.Fatalf("error running the command: %v", err)
		}

		assert.Equal(t, expectedConfigPath, opts.ConfigPath)
		userHomeDir, err := os.UserHomeDir()

		if err != nil {
			t.Fatalf("error getting user cache dir: %v", err)
		}
		assert.Equal(t, filepath.Join(userHomeDir, ".nvimim"), opts.Path)
	})

	t.Run("should create paths if does not exists", func(t *testing.T) {
		var opts AppOptions
		dir, err := os.MkdirTemp("", "nvimim-test-")
		os.Remove(dir)
		// defer os.Remove(dir)
		configDir, err := os.MkdirTemp("", "nvimim-config-test-")
		os.Remove(configDir)
		// defer os.Remove(configDir)
		config_file_name := "nvimim.yaml"
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		os.Setenv("NVIMIM_CONFIG_DIR", configDir)
		os.Setenv("NVIMIM_CONFIG_FILE_NAME", config_file_name)
		os.Setenv("NVIMIM_PATH", dir)
		defer os.Unsetenv("NVIMIM_CONFIG_DIR")
		defer os.Unsetenv("NVIMIM_CONFIG_FILE_NAME")
		defer os.Unsetenv("NVIMIM_PATH")

		parser := flags.NewParser(&opts, flags.Default)
		parser.Usage = "[Application|Help Options] command"

		parser.CommandHandler = WithAppOptions(&opts, WithPathsResolved)

		cmd := &TestOptionsCommand{}

		parser.AddCommand(
			"options",
			"Run test command to check options",
			"Run test command to check options",
			cmd)

		_, err = parser.ParseArgs([]string{"options"})
		if err != nil {
			t.Fatalf("error running the command: %v", err)
		}

		assert.DirExists(t, opts.ConfigDir)
		assert.FileExists(t, opts.ConfigPath)
		assert.DirExists(t, opts.Path)
		assert.DirExists(t, opts.CachePath)
	})
}

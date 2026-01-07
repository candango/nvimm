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
		os.Setenv("NVIMM_CONFIG_DIR", "/etc/nvimm")
		os.Setenv("NVIMM_CONFIG_FILE_NAME", "nvimm.yaml")
		os.Setenv("NVIMM_PATH", "/opt/nvimm")
		os.Setenv("NVIMM_CACHE_PATH", "/opt/nvim/cache")
		defer os.Unsetenv("NVIMM_CONFIG_DIR")
		defer os.Unsetenv("NVIMM_CONFIG_FILE_NAME")
		defer os.Unsetenv("NVIMM_PATH")
		defer os.Unsetenv("NVIMM_CACHE_SUB_DIR")

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

		assert.Equal(t, filepath.Join(os.Getenv("NVIMM_CONFIG_DIR"),
			os.Getenv("NVIMM_CONFIG_FILE_NAME")), opts.ConfigPath)
		assert.Equal(t, os.Getenv("NVIMM_PATH"), opts.Path)
		assert.Equal(t, opts.CachePath, opts.CachePath)
	})

	t.Run("should get default values", func(t *testing.T) {
		var opts AppOptions

		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			t.Fatalf("error getting user config dir: %v", err)
		}
		expectedConfigPath := filepath.Join(userConfigDir, "nvimm", "nvimm.yml")
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
		assert.Equal(t, filepath.Join(userHomeDir, ".nvimm"), opts.Path)
	})

	t.Run("should create paths if does not exists", func(t *testing.T) {
		var opts AppOptions
		dir, err := os.MkdirTemp("", "nvimm-test-")
		os.Remove(dir)
		// defer os.Remove(dir)
		configDir, err := os.MkdirTemp("", "nvimm-config-test-")
		os.Remove(configDir)
		// defer os.Remove(configDir)
		config_file_name := "nvimm.yaml"
		if err != nil {
			t.Fatalf("failed to create temp dir: %v", err)
		}
		os.Setenv("NVIMM_CONFIG_DIR", configDir)
		os.Setenv("NVIMM_CONFIG_FILE_NAME", config_file_name)
		os.Setenv("NVIMM_PATH", dir)
		defer os.Unsetenv("NVIMM_CONFIG_DIR")
		defer os.Unsetenv("NVIMM_CONFIG_FILE_NAME")
		defer os.Unsetenv("NVIMM_PATH")

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

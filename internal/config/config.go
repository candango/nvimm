package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/candango/iook/pathx"
	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_NVIMIM_DIR         = "nvimim"
	DEFAULT_NVIMIM_CONFIG_FILE = "nvimim.yml"
)

// Config holds all the configuration settings
type Config struct {
	CacheDir string        `yaml:"cache_dir"`
	CacheTTL time.Duration `yaml:"cache_ttl"`
	Repo     string        `yaml:"repo"`
}

// NewDefaultConfig returns a Config initialized with standard default values.
func NewDefaultConfig() (*Config, error) {
	userCache, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}
	return &Config{
		CacheDir: filepath.Join(userCache, DEFAULT_NVIMIM_DIR),
		CacheTTL: 24 * time.Hour,
	}, nil
}

// Manager handles the lifecycle of the application configuraion
type Manager struct {
	configPath string
	*Config
}

func NewManager() (*Manager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	return &Manager{
		configPath: filepath.Join(configDir, DEFAULT_NVIMIM_DIR,
			DEFAULT_NVIMIM_CONFIG_FILE),
	}, nil
}

// Load reads the config from disk if it exists, merging it with defaults.
// If the file does not exist, it uses the default settings.
func (m *Manager) Load() error {
	if !pathx.Exists(m.configPath) {
		return fmt.Errorf("config file %s does not exists", m.configPath)
	}
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, m.Config)
}

// Save persists the current configuration to the config.yaml file.
func (m *Manager) Save() error {
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(m.Config)
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

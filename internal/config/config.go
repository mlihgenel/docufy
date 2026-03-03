package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig uygulama yapılandırmasını tutar
type AppConfig struct {
	FirstRunCompleted bool   `json:"first_run_completed"`
	DefaultOutputDir  string `json:"default_output_dir,omitempty"`
}

const (
	configDirName       = ".docufy"
	legacyConfigDirName = ".fileconverter"
)

var userHomeDir = os.UserHomeDir

// configDir yapılandırma dizinini döner (~/.docufy)
func configDir() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, configDirName), nil
}

func legacyConfigDir() (string, error) {
	home, err := userHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, legacyConfigDirName), nil
}

// configPath yapılandırma dosya yolunu döner
func configPath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func legacyConfigPath() (string, error) {
	dir, err := legacyConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

func configReadPaths() []string {
	paths := make([]string, 0, 2)
	if path, err := configPath(); err == nil {
		paths = append(paths, path)
	}
	if path, err := legacyConfigPath(); err == nil {
		paths = append(paths, path)
	}
	return paths
}

// LoadConfig yapılandırmayı dosyadan okur
func LoadConfig() (*AppConfig, error) {
	for _, path := range configReadPaths() {
		data, err := os.ReadFile(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return &AppConfig{}, nil
		}

		var cfg AppConfig
		if err := json.Unmarshal(data, &cfg); err != nil {
			return &AppConfig{}, nil
		}

		return &cfg, nil
	}
	return &AppConfig{}, nil
}

// SaveConfig yapılandırmayı dosyaya kaydeder
func SaveConfig(cfg *AppConfig) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(dir, "config.json")
	return os.WriteFile(path, data, 0644)
}

// IsFirstRun uygulamanın ilk kez çalıştırılıp çalıştırılmadığını kontrol eder
func IsFirstRun() bool {
	cfg, _ := LoadConfig()
	return !cfg.FirstRunCompleted
}

// MarkFirstRunDone ilk çalıştırma tamamlandı olarak işaretler
func MarkFirstRunDone() error {
	cfg, _ := LoadConfig()
	cfg.FirstRunCompleted = true
	return SaveConfig(cfg)
}

// GetDefaultOutputDir varsayılan çıktı dizinini döner
func GetDefaultOutputDir() string {
	cfg, _ := LoadConfig()
	return cfg.DefaultOutputDir
}

// SetDefaultOutputDir varsayılan çıktı dizinini kaydeder
func SetDefaultOutputDir(dir string) error {
	cfg, _ := LoadConfig()
	cfg.DefaultOutputDir = dir
	return SaveConfig(cfg)
}

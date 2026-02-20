package pages

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const configFileName = "day1.yml"

type Brand struct {
	Name string `yaml:"name"`
	Logo string `yaml:"logo"`
}

type Config struct {
	Brand       Brand    `yaml:"brand"`
	HelpURL     string   `yaml:"help_url"`
	Theme       string   `yaml:"theme"`
	Title       string   `yaml:"title"`
	AccentColor string   `yaml:"accent_color"`
	FinalPage   string   `yaml:"final_page"`
	Pages       []string `yaml:"pages"`
}

// LoadConfig reads day1.yml from pagesDir. Returns zero Config if the file
// doesn't exist.
func LoadConfig(pagesDir string) (Config, error) {
	data, err := os.ReadFile(filepath.Join(pagesDir, configFileName))
	if os.IsNotExist(err) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

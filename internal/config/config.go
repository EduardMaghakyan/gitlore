package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Distill DistillConfig `toml:"distill"`
	Notes   NotesConfig   `toml:"notes"`
}

type DistillConfig struct {
	UseCLI bool   `toml:"use_cli"`
	Model  string `toml:"model"`
	APIKey string `toml:"api_key"`
}

type NotesConfig struct {
	AutoPush bool `toml:"auto_push"`
}

func defaults() *Config {
	return &Config{
		Distill: DistillConfig{
			UseCLI: true,
			Model:  "claude-sonnet-4-6",
		},
		Notes: NotesConfig{
			AutoPush: false,
		},
	}
}

func Load(repoRoot string) (*Config, error) {
	cfg := defaults()

	// Global config
	home, err := os.UserHomeDir()
	if err == nil {
		loadFile(filepath.Join(home, ".gitlore"), cfg)
	}

	// Repo config (overrides global)
	if repoRoot != "" {
		loadFile(filepath.Join(repoRoot, ".gitlore"), cfg)
	}

	// Environment overrides
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		cfg.Distill.APIKey = key
	}
	if model := os.Getenv("GITLORE_MODEL"); model != "" {
		cfg.Distill.Model = model
	}

	return cfg, nil
}

func loadFile(path string, cfg *Config) {
	if _, err := os.Stat(path); err != nil {
		return
	}
	toml.DecodeFile(path, cfg)
}

package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const DefaultModel = "claude-sonnet-4-20250514"

// Config holds declog configuration.
type Config struct {
	AI AIConfig
}

// AIConfig holds AI-related configuration.
type AIConfig struct {
	Model string
}

// Load reads the config from ~/.config/declog/config.toml.
// Returns a default Config if the file does not exist.
func Load() (*Config, error) {
	path := configPath()
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return &Config{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("opening config: %w", err)
	}
	defer f.Close()
	return parse(f)
}

// ResolveModel returns the model to use, applying priority:
// flagModel > config file model > default.
func ResolveModel(flagModel string, cfg *Config) string {
	if flagModel != "" {
		return flagModel
	}
	if cfg != nil && cfg.AI.Model != "" {
		return cfg.AI.Model
	}
	return DefaultModel
}

func configPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "declog", "config.toml")
}

func parse(r io.Reader) (*Config, error) {
	cfg := &Config{}
	scanner := bufio.NewScanner(r)
	inAI := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if line == "[ai]" {
			inAI = true
			continue
		}
		if strings.HasPrefix(line, "[") {
			inAI = false
			continue
		}
		if inAI && strings.HasPrefix(line, "model") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				cfg.AI.Model = strings.Trim(strings.TrimSpace(parts[1]), `"`)
			}
		}
	}
	return cfg, scanner.Err()
}

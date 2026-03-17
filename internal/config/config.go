package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const DefaultModel = "claude-sonnet-4-20250514"

// Config holds declog configuration.
type Config struct {
	AI        AIConfig
	Decisions DecisionsConfig
	Review    ReviewConfig
	Editor    EditorConfig
}

// AIConfig holds AI-related configuration.
type AIConfig struct {
	Model string
}

// DecisionsConfig holds decisions directory configuration.
type DecisionsConfig struct {
	Dir string
}

// ReviewConfig holds review threshold configuration.
type ReviewConfig struct {
	MonthsThreshold int
}

// EditorConfig holds editor configuration.
type EditorConfig struct {
	Command string
}

// Load reads the config from ~/.config/kizami/config.toml.
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
	var section string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = line[1 : len(line)-1]
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.Trim(strings.TrimSpace(parts[1]), `"`)
		switch section {
		case "ai":
			if key == "model" {
				cfg.AI.Model = val
			}
		case "decisions":
			if key == "dir" {
				cfg.Decisions.Dir = val
			}
		case "review":
			if key == "months_threshold" {
				n, _ := strconv.Atoi(val)
				cfg.Review.MonthsThreshold = n
			}
		case "editor":
			if key == "command" {
				cfg.Editor.Command = val
			}
		}
	}
	return cfg, scanner.Err()
}

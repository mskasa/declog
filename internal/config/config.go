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

// Config holds kizami configuration.
type Config struct {
	AI        AIConfig
	Documents DocumentsConfig
	Decisions DecisionsConfig
	Design    DesignConfig
	Audit     AuditConfig
	Review    ReviewConfig
	Editor    EditorConfig
}

// AIConfig holds AI-related configuration.
type AIConfig struct {
	Model string
}

// DocumentsConfig holds the list of document directories for commands like
// list, search, show, blame, review, status, and supersede.
type DocumentsConfig struct {
	Dirs []string
}

// DecisionsConfig holds decisions directory configuration.
type DecisionsConfig struct {
	Dir string
}

// DesignConfig holds design documents directory configuration.
type DesignConfig struct {
	Dir string
}

// AuditConfig holds audit directory configuration.
type AuditConfig struct {
	Dirs []string
}

// ReviewConfig holds review threshold configuration.
type ReviewConfig struct {
	MonthsThreshold int
}

// EditorConfig holds editor configuration.
type EditorConfig struct {
	Command string
}

// Load reads the config for the given project root.
// It first looks for kizami.toml in root, then falls back to ~/.config/kizami/config.toml.
// Returns a default Config if neither file exists.
func Load(root string) (*Config, error) {
	if root != "" {
		projectPath := filepath.Join(root, "kizami.toml")
		f, err := os.Open(projectPath)
		if err == nil {
			defer f.Close()
			return parse(f)
		}
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("opening config: %w", err)
		}
	}

	path := globalConfigPath()
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

func globalConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "kizami", "config.toml")
}

// parseStringArray parses a TOML inline string array like ["a", "b", "c"].
func parseStringArray(raw string) []string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "[")
	raw = strings.TrimSuffix(raw, "]")
	var result []string
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		item = strings.Trim(item, `"'`)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
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
		case "documents":
			if key == "dirs" {
				cfg.Documents.Dirs = parseStringArray(parts[1])
			}
		case "decisions":
			if key == "dir" {
				cfg.Decisions.Dir = val
			}
		case "design":
			if key == "dir" {
				cfg.Design.Dir = val
			}
		case "audit":
			if key == "dirs" {
				cfg.Audit.Dirs = parseStringArray(parts[1])
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

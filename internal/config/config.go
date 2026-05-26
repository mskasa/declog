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

// EnvModel is the environment variable kizami reads for the model name,
// following the same convention as Claude Code (CLAUDE_CODE_USE_BEDROCK=1).
const EnvModel = "ANTHROPIC_MODEL"

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
	Model   string
	Backend string // "anthropic" (default) or "bedrock"
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
// If kizami.local.toml exists in root, its non-zero values override the base config.
// Returns a default Config if neither file exists.
func Load(root string) (*Config, error) {
	var cfg *Config
	var err error

	if root != "" {
		projectPath := filepath.Join(root, "kizami.toml")
		f, ferr := os.Open(projectPath)
		if ferr == nil {
			defer f.Close()
			cfg, err = parse(f)
			if err != nil {
				return nil, err
			}
		} else if !os.IsNotExist(ferr) {
			return nil, fmt.Errorf("opening config: %w", ferr)
		}
	}

	if cfg == nil {
		path := globalConfigPath()
		f, ferr := os.Open(path)
		if os.IsNotExist(ferr) {
			cfg = &Config{}
		} else if ferr != nil {
			return nil, fmt.Errorf("opening config: %w", ferr)
		} else {
			defer f.Close()
			cfg, err = parse(f)
			if err != nil {
				return nil, err
			}
		}
	}

	if root != "" {
		localPath := filepath.Join(root, "kizami.local.toml")
		f, ferr := os.Open(localPath)
		if ferr == nil {
			defer f.Close()
			local, lerr := parse(f)
			if lerr != nil {
				return nil, fmt.Errorf("opening local config: %w", lerr)
			}
			merge(cfg, local)
		} else if !os.IsNotExist(ferr) {
			return nil, fmt.Errorf("opening local config: %w", ferr)
		}
	}

	return cfg, nil
}

// merge overlays non-zero values from local into base.
func merge(base, local *Config) {
	if local.AI.Model != "" {
		base.AI.Model = local.AI.Model
	}
	if local.AI.Backend != "" {
		base.AI.Backend = local.AI.Backend
	}
	if len(local.Documents.Dirs) > 0 {
		base.Documents.Dirs = local.Documents.Dirs
	}
	if local.Decisions.Dir != "" {
		base.Decisions.Dir = local.Decisions.Dir
	}
	if local.Design.Dir != "" {
		base.Design.Dir = local.Design.Dir
	}
	if len(local.Audit.Dirs) > 0 {
		base.Audit.Dirs = local.Audit.Dirs
	}
	if local.Review.MonthsThreshold != 0 {
		base.Review.MonthsThreshold = local.Review.MonthsThreshold
	}
	if local.Editor.Command != "" {
		base.Editor.Command = local.Editor.Command
	}
}

// ResolveModel returns the model to use, applying priority:
// flagModel > ANTHROPIC_MODEL env var > config file model > default.
func ResolveModel(flagModel string, cfg *Config) string {
	if flagModel != "" {
		return flagModel
	}
	if env := os.Getenv(EnvModel); env != "" {
		return env
	}
	if cfg != nil && cfg.AI.Model != "" {
		return cfg.AI.Model
	}
	return DefaultModel
}

// IsBedrockModel reports whether the model ID indicates an AWS Bedrock model.
// Recognised patterns:
//   - ARN: "arn:aws:bedrock:..."  (Application/Provisioned Inference Profile)
//   - Cross-region inference profile: "us.*", "eu.*", "ap.*"
//   - Standard Bedrock model ID: provider-prefixed format such as "anthropic.claude-*", "amazon.nova-*"
func IsBedrockModel(model string) bool {
	if strings.HasPrefix(model, "arn:aws:bedrock:") {
		return true
	}
	for _, prefix := range []string{"us.", "eu.", "ap."} {
		if strings.HasPrefix(model, prefix) {
			return true
		}
	}
	for _, provider := range []string{"anthropic.", "amazon.", "meta.", "mistral.", "cohere.", "ai21.", "stability."} {
		if strings.HasPrefix(model, provider) {
			return true
		}
	}
	return false
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
			switch key {
			case "model":
				cfg.AI.Model = val
			case "backend":
				cfg.AI.Backend = val
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

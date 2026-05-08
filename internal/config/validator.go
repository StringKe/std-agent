package config

import (
	"fmt"
	"os"
)

// Validate 校验 Config schema 并做轻量 normalize（如 BackupKeep 下限）
func Validate(cfg *Config) error {
	if cfg.Version == "" {
		return fmt.Errorf("config: version is required")
	}
	if cfg.Version != "1.0" {
		return fmt.Errorf("config: unsupported schema version %q (expected 1.0)", cfg.Version)
	}

	for name := range cfg.Targets {
		if !IsValidTarget(name) {
			return fmt.Errorf("config: unknown target %q (valid: %v)", name, ValidTargets)
		}
	}

	for name, src := range cfg.Sources {
		if src.URL == "" {
			return fmt.Errorf("config: source %q missing url", name)
		}
		if len(src.Paths) == 0 {
			return fmt.Errorf("config: source %q missing paths", name)
		}
		switch src.Auth {
		case "", "none", "ssh", "https-token":
		default:
			return fmt.Errorf("config: source %q invalid auth %q", name, src.Auth)
		}
		if src.Auth == "https-token" {
			if src.TokenEnv == "" {
				return fmt.Errorf("config: source %q https-token requires token_env", name)
			}
			if os.Getenv(src.TokenEnv) == "" {
				return fmt.Errorf("config: source %q token_env %q is empty", name, src.TokenEnv)
			}
		}
	}

	if cfg.BackupKeep < 1 {
		cfg.BackupKeep = 1
	}

	return nil
}

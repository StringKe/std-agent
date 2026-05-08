package config

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Load 从 path 加载 .stdai/config.toml 并应用环境变量覆盖与校验
func Load(path string) (*Config, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // CLI 用户显式指定的配置路径
	if err != nil {
		return nil, fmt.Errorf("load config %s: %w", path, err)
	}
	var cfg Config
	if err := toml.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	applyEnvOverrides(&cfg)
	if err := Validate(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// applyEnvOverrides 把 STDAI_* 环境变量映射到 Config 字段
func applyEnvOverrides(cfg *Config) {
	if os.Getenv("STDAI_DRY_RUN") == "1" {
		cfg.DryRun = true
	}
	if os.Getenv("STDAI_VERBOSE") == "1" {
		cfg.Verbose = true
	}
	if os.Getenv("STDAI_NO_PULL") == "1" {
		cfg.AutoPull = false
	}
}

// Save 把 Config 写到 path（toml 格式），仅 init 命令使用
func Save(path string, cfg *Config) error {
	raw, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
	}
	return nil
}

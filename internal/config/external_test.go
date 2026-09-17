package config

import (
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
)

// TestExternalEnabledDefaults [external] 缺省（老配置）视为启用
func TestExternalEnabledDefaults(t *testing.T) {
	var nilCfg *Config
	if !nilCfg.ExternalEnabled() {
		t.Error("nil config should default external to enabled")
	}
	cfg := &Config{}
	if !cfg.ExternalEnabled() {
		t.Error("missing [external] table should default to enabled")
	}
	cfg.External = &ExternalConfig{Enabled: false}
	if cfg.ExternalEnabled() {
		t.Error("explicit enabled=false should disable")
	}
	if !Default().ExternalEnabled() {
		t.Error("Default() should enable external adopt")
	}
}

// TestExternalRoundTrip 默认配置 marshal 后 [external] 表仍可解析且启用
func TestExternalRoundTrip(t *testing.T) {
	raw, err := toml.Marshal(Default())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), "[external]") {
		t.Errorf("marshaled config should contain [external] table:\n%s", raw)
	}
	var back Config
	if err := toml.Unmarshal(raw, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := Validate(&back); err != nil {
		t.Fatalf("validate round-trip: %v", err)
	}
	if !back.ExternalEnabled() {
		t.Error("round-trip should keep external enabled")
	}
}

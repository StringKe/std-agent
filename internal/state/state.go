package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

// StateFile 是 state.json 的默认相对路径
const StateFile = ".stdai/state.json"

// State 是 sync 之间持久化的状态
type State struct {
	Version  string            `json:"version"`
	LastSync time.Time         `json:"last_sync"`
	Sources  map[string]Source `json:"sources,omitempty"`
	Targets  map[string]Target `json:"targets,omitempty"`
}

// Source 是单个源的同步状态
type Source struct {
	LastPull time.Time `json:"last_pull"`
	Commit   string    `json:"commit,omitempty"`
}

// Target 是单 target 的输出快照
type Target struct {
	LastSync time.Time         `json:"last_sync"`
	Outputs  map[string]string `json:"outputs,omitempty"` // path -> sha256 hex
}

// Load 从 path 加载 state；不存在返回 zero State
func Load(path string) (*State, error) {
	raw, err := os.ReadFile(path) //nolint:gosec // CLI 用户已知路径
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &State{Version: "1.0"}, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, fmt.Errorf("parse state %s: %w", path, err)
	}
	if s.Version == "" {
		s.Version = "1.0"
	}
	return &s, nil
}

// Save 把 state 写到 path（atomic）
func Save(path string, s *State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

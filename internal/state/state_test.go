package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadMissing(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Version != "1.0" {
		t.Errorf("version = %s, want 1.0", s.Version)
	}
}

func TestStateAccumulatesAcrossSyncs(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "state.json")

	now := time.Now().UTC().Truncate(time.Second)

	// 第 1 次 sync 写入 claude-code
	s1 := &State{
		Version:  "1.0",
		LastSync: now,
		Targets: map[string]Target{
			"claude-code": {
				LastSync: now,
				Outputs:  map[string]string{"CLAUDE.md": "hash1"},
			},
		},
	}
	if err := Save(path, s1); err != nil {
		t.Fatal(err)
	}

	// 第 2 次 sync：reload + 加 codex target
	s2, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s2.Targets == nil {
		s2.Targets = map[string]Target{}
	}
	s2.Targets["codex"] = Target{
		LastSync: now.Add(time.Hour),
		Outputs:  map[string]string{"AGENTS.md": "hash2"},
	}
	if err := Save(path, s2); err != nil {
		t.Fatal(err)
	}

	// 第 3 次 reload 应当含 2 个 target
	s3, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(s3.Targets) != 2 {
		t.Errorf("expected 2 targets after second sync, got %d", len(s3.Targets))
	}
	if s3.Targets["claude-code"].Outputs["CLAUDE.md"] != "hash1" {
		t.Error("first sync output should persist")
	}
	if s3.Targets["codex"].Outputs["AGENTS.md"] != "hash2" {
		t.Error("second sync output should be added")
	}
}

func TestStateLoadCorruptJSON(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "state.json")
	if err := os.WriteFile(path, []byte("not valid json"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(path); err == nil {
		t.Error("expected error on corrupt JSON")
	}
}

func TestStateSaveCreatesParentDir(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "deep", "nested", "state.json")
	s := &State{Version: "1.0"}
	if err := Save(path, s); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("file not created: %v", err)
	}
}

func TestStateLoadEmptyVersionDefaults(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "state.json")
	// 写无 version 字段的 json
	if err := os.WriteFile(path, []byte(`{"last_sync":"2026-01-01T00:00:00Z"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Version != "1.0" {
		t.Errorf("Version should default to '1.0', got %q", s.Version)
	}
}

func TestStateOverwriteAtomic(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "state.json")
	s1 := &State{Version: "1.0", LastSync: time.Now()}
	if err := Save(path, s1); err != nil {
		t.Fatal(err)
	}
	s2 := &State{Version: "1.0", LastSync: time.Now().Add(time.Hour)}
	if err := Save(path, s2); err != nil {
		t.Fatalf("overwrite Save failed: %v", err)
	}
	// .tmp 文件不应残留
	if _, err := os.Stat(path + ".tmp"); err == nil {
		t.Error(".tmp file should be removed after rename")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "state.json")

	now := time.Date(2026, 5, 8, 12, 0, 0, 0, time.UTC)
	s := &State{
		Version:  "1.0",
		LastSync: now,
		Targets: map[string]Target{
			"claude-code": {
				LastSync: now,
				Outputs:  map[string]string{"CLAUDE.md": "sha256-hex"},
			},
		},
	}
	if err := Save(path, s); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.Version != "1.0" {
		t.Errorf("version = %s", got.Version)
	}
	if !got.LastSync.Equal(now) {
		t.Errorf("LastSync = %v, want %v", got.LastSync, now)
	}
	if got.Targets["claude-code"].Outputs["CLAUDE.md"] != "sha256-hex" {
		t.Errorf("output mismatch: %v", got.Targets)
	}
}

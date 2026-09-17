package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportCommandAdoptsToDisk(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	setupSyncProject(t, tmp)
	t.Chdir(tmp)

	mustMkdirAll(t, filepath.Join(tmp, ".ai/guidelines"))
	mustWriteFile(t, filepath.Join(tmp, ".ai/guidelines/ext.md"), "External body.\n")

	cmd := newImportCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("import: %v", err)
	}
	if !strings.Contains(out.String(), "[import]") {
		t.Errorf("missing [import] in output:\n%s", out.String())
	}
	disk, err := os.ReadFile(filepath.Join(tmp, ".stdai/standards/external/guidelines/ext.md"))
	if err != nil {
		t.Fatalf("adopted file not on disk: %v", err)
	}
	if !strings.Contains(string(disk), "External body.") || !strings.Contains(string(disk), "external_source") {
		t.Errorf("adopted file wrong:\n%s", disk)
	}

	// Second run is idempotent: no new files reported
	cmd2 := newImportCmd()
	var out2 bytes.Buffer
	cmd2.SetOut(&out2)
	cmd2.SetErr(new(bytes.Buffer))
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("second import: %v", err)
	}
	if !strings.Contains(out2.String(), "no external artifacts") {
		t.Errorf("second import should report nothing new:\n%s", out2.String())
	}
}

func TestSyncReportsExternalCount(t *testing.T) {
	old := flagConfig
	defer func() { flagConfig = old }()
	flagConfig = ".stdai/config.toml"

	tmp := t.TempDir()
	setupSyncProject(t, tmp)
	t.Chdir(tmp)

	mustMkdirAll(t, filepath.Join(tmp, ".ai/guidelines"))
	mustWriteFile(t, filepath.Join(tmp, ".ai/guidelines/ext.md"), "External body.\n")

	cmd := newSyncCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(new(bytes.Buffer))
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !strings.Contains(out.String(), "[external]") {
		t.Errorf("sync should report [external] count:\n%s", out.String())
	}

	// --no-external: skip auto-adopt on a fresh project (nothing written, no count line)
	tmp2 := t.TempDir()
	setupSyncProject(t, tmp2)
	t.Chdir(tmp2)
	mustMkdirAll(t, filepath.Join(tmp2, ".ai/guidelines"))
	mustWriteFile(t, filepath.Join(tmp2, ".ai/guidelines/ext.md"), "External body.\n")

	cmd2 := newSyncCmd()
	var out2 bytes.Buffer
	cmd2.SetOut(&out2)
	cmd2.SetErr(new(bytes.Buffer))
	cmd2.SetArgs([]string{"--no-external"})
	if err := cmd2.Execute(); err != nil {
		t.Fatalf("sync --no-external: %v", err)
	}
	if strings.Contains(out2.String(), "[external]") {
		t.Errorf("--no-external should suppress [external] line:\n%s", out2.String())
	}
	if _, err := os.Stat(filepath.Join(tmp2, ".stdai/standards/external")); !os.IsNotExist(err) {
		t.Error("--no-external must not materialize external/")
	}
}

package source

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// setupBareWithCommit 创建一个 bare upstream + working repo 推一个 commit
func setupBareWithCommit(t *testing.T, files map[string]string) string {
	t.Helper()

	bare := t.TempDir()
	if _, err := git.PlainInit(bare, true); err != nil {
		t.Fatalf("PlainInit bare: %v", err)
	}

	work := t.TempDir()
	repo, err := git.PlainInit(work, false)
	if err != nil {
		t.Fatalf("PlainInit work: %v", err)
	}

	for rel, content := range files {
		full := filepath.Join(work, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	w, err := repo.Worktree()
	if err != nil {
		t.Fatal(err)
	}
	for rel := range files {
		if _, err := w.Add(rel); err != nil {
			t.Fatalf("Add %s: %v", rel, err)
		}
	}

	sig := &object.Signature{Name: "test", Email: "t@e.com", When: time.Now()}
	if _, err := w.Commit("init", &git.CommitOptions{Author: sig, Committer: sig}); err != nil {
		t.Fatalf("Commit: %v", err)
	}

	if _, err := repo.CreateRemote(&gitconfig.RemoteConfig{
		Name: "origin",
		URLs: []string{bare},
	}); err != nil {
		t.Fatalf("CreateRemote: %v", err)
	}
	if err := repo.Push(&git.PushOptions{RemoteName: "origin"}); err != nil {
		t.Fatalf("Push: %v", err)
	}
	return bare
}

func TestGitPullCloneAndFiles(t *testing.T) {
	bare := setupBareWithCommit(t, map[string]string{
		"standards/rules/x.md":                "---\ntype: rules\nname: x\n---\nbody\n",
		"standards/skills/foo/SKILL.md":       "---\ntype: skills\nname: foo\n---\nbody\n",
		"standards/skills/foo/scripts/run.sh": "#!/bin/sh\necho ok\n",
		"README.md":                           "# proj",
	})

	cacheParent := t.TempDir()
	cacheDir := filepath.Join(cacheParent, "default")
	g := &Git{
		NameValue: "default",
		URL:       bare,
		Branch:    "master",
		CacheDir:  cacheDir,
		Paths:     []string{"standards/"},
	}

	if err := g.Pull(); err != nil {
		t.Fatalf("Pull: %v", err)
	}
	if g.Name() != "default" {
		t.Errorf("Name = %s", g.Name())
	}

	files, err := g.Files()
	if err != nil {
		t.Fatal(err)
	}
	got := make(map[string]bool, len(files))
	for _, f := range files {
		got[f.Path] = true
	}
	for _, want := range []string{
		"rules/x.md",
		"skills/foo/SKILL.md",
		"skills/foo/scripts/run.sh",
	} {
		if !got[want] {
			t.Errorf("missing %s in %v", want, got)
		}
	}
	// README.md 在 standards/ 之外，不应出现
	if got["README.md"] {
		t.Error("README.md outside standards/ should not be included")
	}

	// 第二次 Pull 应该是 no-op（NoErrAlreadyUpToDate 被吞掉）
	if err := g.Pull(); err != nil {
		t.Errorf("second Pull (already up-to-date) failed: %v", err)
	}
}

func TestGitFilesNoPaths(t *testing.T) {
	bare := setupBareWithCommit(t, map[string]string{
		"foo.md": "body",
	})
	cacheDir := filepath.Join(t.TempDir(), "x")
	g := &Git{
		NameValue: "x",
		URL:       bare,
		Branch:    "master",
		CacheDir:  cacheDir,
		Paths:     nil, // 空 -> 整个 cacheDir
	}
	if err := g.Pull(); err != nil {
		t.Fatal(err)
	}
	files, err := g.Files()
	if err != nil {
		t.Fatal(err)
	}
	got := false
	for _, f := range files {
		if f.Path == "foo.md" {
			got = true
		}
	}
	if !got {
		t.Errorf("expected foo.md, got %v", files)
	}
}

func TestGitBuildAuthVariants(t *testing.T) {
	cases := []struct {
		auth     string
		tokenEnv string
		envVal   string
		wantErr  bool
		wantNil  bool
	}{
		{"", "", "", false, true},
		{"none", "", "", false, true},
		{"https-token", "STDAGENT_TEST_TOKEN_X", "fake", false, false},
		{"https-token", "STDAGENT_TEST_TOKEN_MISSING", "", true, false},
		{"weird", "", "", true, false},
	}
	for _, c := range cases {
		t.Run(c.auth+"_"+c.tokenEnv, func(t *testing.T) {
			if c.envVal != "" {
				t.Setenv(c.tokenEnv, c.envVal)
			} else if c.tokenEnv != "" {
				_ = os.Unsetenv(c.tokenEnv)
			}
			g := &Git{Auth: c.auth, TokenEnv: c.tokenEnv}
			auth, err := g.buildAuth()
			if c.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if c.wantNil && auth != nil {
				t.Errorf("expected nil auth, got %v", auth)
			}
			if !c.wantNil && auth == nil {
				t.Error("expected non-nil auth")
			}
		})
	}
}

package source

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
)

// Git 是 git 远端源
type Git struct {
	NameValue string
	URL       string
	Branch    string
	CacheDir  string
	Paths     []string
	Auth      string
	TokenEnv  string
}

// Name 返回配置中的 source 名
func (g *Git) Name() string { return g.NameValue }

// Pull 克隆或拉取最新
func (g *Git) Pull() error {
	auth, err := g.buildAuth()
	if err != nil {
		return err
	}
	branch := g.Branch
	if branch == "" {
		branch = "main"
	}

	if _, statErr := os.Stat(filepath.Join(g.CacheDir, ".git")); errors.Is(statErr, os.ErrNotExist) {
		if err := os.MkdirAll(g.CacheDir, 0o750); err != nil {
			return err
		}
		_, err := git.PlainClone(g.CacheDir, false, &git.CloneOptions{
			URL:           g.URL,
			ReferenceName: plumbing.NewBranchReferenceName(branch),
			SingleBranch:  true,
			Depth:         1,
			Auth:          auth,
		})
		if err != nil {
			return fmt.Errorf("clone %s: %w", g.URL, err)
		}
		return nil
	}

	repo, err := git.PlainOpen(g.CacheDir)
	if err != nil {
		return fmt.Errorf("open cache %s: %w", g.CacheDir, err)
	}
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	if err := w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
		Auth:          auth,
		Force:         true,
	}); err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return fmt.Errorf("pull %s: %w", g.URL, err)
	}
	return nil
}

// Files 扫描 cache + paths 下所有 .md
func (g *Git) Files() ([]File, error) {
	var out []File
	if len(g.Paths) == 0 {
		l := &Local{Root: g.CacheDir}
		return l.Files()
	}
	for _, p := range g.Paths {
		root := filepath.Join(g.CacheDir, p)
		l := &Local{Root: root}
		files, err := l.Files()
		if err != nil {
			return nil, err
		}
		out = append(out, files...)
	}
	return out, nil
}

// buildAuth 根据 Auth 字段构造 transport 认证
func (g *Git) buildAuth() (transport.AuthMethod, error) {
	switch g.Auth {
	case "", "none":
		return nil, nil
	case "ssh":
		auth, err := gitssh.NewSSHAgentAuth("git")
		if err != nil {
			return nil, fmt.Errorf("ssh agent: %w", err)
		}
		return auth, nil
	case "https-token":
		token := os.Getenv(g.TokenEnv)
		if token == "" {
			return nil, fmt.Errorf("token env %q is empty", g.TokenEnv)
		}
		return &githttp.BasicAuth{Username: "x-access-token", Password: token}, nil
	}
	return nil, fmt.Errorf("unsupported auth %q", g.Auth)
}

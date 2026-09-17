package cli

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/writer"
)

//go:embed init_assets/help/*.md init_assets/root.md
var initAssets embed.FS

func newInitCmd() *cobra.Command {
	var force, minimal bool
	var sourceURL string
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize .stdai/ and config.toml",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(cmd, initOptions{Force: force, Minimal: minimal, Source: sourceURL})
		},
	}
	f := cmd.Flags()
	f.BoolVar(&force, "force", false, "Overwrite an existing .stdai/ (back it up first)")
	f.BoolVar(&minimal, "minimal", false, "Skip example files, create empty dirs only")
	f.StringVar(&sourceURL, "source", "", "Write the given URL into [sources.default]")
	return cmd
}

type initOptions struct {
	Force   bool
	Minimal bool
	Source  string
}

func runInit(cmd *cobra.Command, opts initOptions) error {
	root := "."
	stdaiDir := filepath.Join(root, ".stdai")

	if _, err := os.Stat(stdaiDir); err == nil {
		if !opts.Force {
			return fmt.Errorf(".stdai/ already exists; use --force to overwrite")
		}
		bak := stdaiDir + "-backup-" + time.Now().UTC().Format("20060102T150405Z")
		if rerr := os.Rename(stdaiDir, bak); rerr != nil {
			return rerr
		}
		cmd.Printf("backup: %s -> %s\n", stdaiDir, bak)
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	for _, sub := range []string{
		"standards/rules",
		"standards/skills",
		"standards/commands",
		"standards/subagents",
		"standards/references",
		"help",
		"cache",
		"backups",
		"logs",
	} {
		dir := filepath.Join(stdaiDir, sub)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, ".gitkeep"), []byte{}, 0o600); err != nil {
			return err
		}
	}

	// Ship the built-in stdagent concept docs to .stdai/help/ (excluded from
	// sync; root.md references them via @<path>).
	if err := writeHelpAssets(stdaiDir); err != nil {
		return err
	}

	// Write the example .stdai/standards/root.md (the AI rewrites it with the
	// project's real content on first takeover).
	if err := writeRootTemplate(stdaiDir); err != nil {
		return err
	}

	cfg := config.Default()
	if opts.Source != "" {
		cfg.Sources = map[string]config.SourceConfig{
			"default": {
				URL:     opts.Source,
				Branch:  "main",
				Enabled: true,
				Paths:   []string{"standards/"},
			},
		}
	}
	cfgPath := filepath.Join(stdaiDir, "config.toml")
	if err := config.Save(cfgPath, cfg); err != nil {
		return err
	}

	ignorePath := filepath.Join(root, ".stdaiignore")
	if _, err := os.Stat(ignorePath); errors.Is(err, fs.ErrNotExist) {
		if werr := os.WriteFile(ignorePath, []byte(stdaiIgnoreTemplate), 0o600); werr != nil {
			cmd.Printf("warn: write .stdaiignore: %v\n", werr)
		}
	}

	if err := writeInitGitignore(root, cfg); err != nil {
		cmd.Printf("warn: failed to update .gitignore: %v\n", err)
	}

	cmd.Printf("Initialized .stdai/ at %s\n", stdaiDir)
	cmd.Println("Next steps:")
	cmd.Println("  1. AI: read .stdai/help/stdagent-workflow.md for workflow")
	cmd.Println("  2. AI: rewrite .stdai/standards/root.md with project content")
	cmd.Println("  3. AI: write rules / skills / commands / subagents to .stdai/standards/<type>/")
	cmd.Println("  4. run: stdagent sync")
	return nil
}

// writeHelpAssets copies init_assets/help/*.md to .stdai/help/.
// These are the built-in stdagent concept docs, excluded from sync
// (the runner only scans .stdai/standards/).
func writeHelpAssets(stdaiDir string) error {
	entries, err := initAssets.ReadDir("init_assets/help")
	if err != nil {
		return fmt.Errorf("read embedded help: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, rerr := initAssets.ReadFile("init_assets/help/" + e.Name())
		if rerr != nil {
			return rerr
		}
		dst := filepath.Join(stdaiDir, "help", e.Name())
		if werr := os.WriteFile(dst, data, 0o600); werr != nil {
			return werr
		}
	}
	return nil
}

// writeRootTemplate copies init_assets/root.md to .stdai/standards/root.md.
// On first takeover the AI rewrites this file with the project's real
// content, keeping the stdagent concept reference section.
func writeRootTemplate(stdaiDir string) error {
	data, err := initAssets.ReadFile("init_assets/root.md")
	if err != nil {
		return err
	}
	dst := filepath.Join(stdaiDir, "standards", "root.md")
	return os.WriteFile(dst, data, 0o600)
}

const stdaiIgnoreTemplate = `# .stdaiignore: gitignore-style globs; matched source files are excluded from sync
# Paths are relative to .stdai/standards/; supports doublestar (**) plus ? and * wildcards
# Lines starting with # are comments; blank lines are ignored

# Draft files (example)
# rules/draft-*.md
# **/wip-*.md

# Internal docs never sent to any target
# references/internal-*.md
`

func writeInitGitignore(root string, cfg *config.Config) error {
	if cfg == nil {
		return nil
	}
	mode := config.NormalizeGitignore(cfg.Gitignore)
	if mode == config.GitignoreOff {
		return nil
	}
	var enabled []string
	for name, t := range cfg.Targets {
		if t.Enabled {
			enabled = append(enabled, name)
		}
	}
	_, err := writer.UpsertGitignore(root, writer.GitignoreEntries(mode, enabled, nil), false)
	return err
}

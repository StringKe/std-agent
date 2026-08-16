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
		Short: "初始化 .stdai/ 与 config.toml",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(cmd, initOptions{Force: force, Minimal: minimal, Source: sourceURL})
		},
	}
	f := cmd.Flags()
	f.BoolVar(&force, "force", false, "已存在 .stdai/ 时强制覆盖（先备份）")
	f.BoolVar(&minimal, "minimal", false, "不写示例文件，只建空目录")
	f.StringVar(&sourceURL, "source", "", "在 [sources.default] 写入指定 URL")
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

	// 分发内置 stdagent 概念文档到 .stdai/help/（不参与 sync，root.md 通过 @<path> 引用）
	if err := writeHelpAssets(stdaiDir); err != nil {
		return err
	}

	// 写示例 .stdai/standards/root.md（AI 接管时按项目实际内容重写）
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

// writeHelpAssets 把 init_assets/help/*.md 拷贝到 .stdai/help/。
// 这些是 stdagent 自带的概念解释文档，不参与 sync（runner 只扫 .stdai/standards/）。
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

// writeRootTemplate 把 init_assets/root.md 拷贝到 .stdai/standards/root.md。
// AI 第一次接管时应当根据项目实际内容重写本文件，保留 stdagent 概念引用段落。
func writeRootTemplate(stdaiDir string) error {
	data, err := initAssets.ReadFile("init_assets/root.md")
	if err != nil {
		return err
	}
	dst := filepath.Join(stdaiDir, "standards", "root.md")
	return os.WriteFile(dst, data, 0o600)
}

const stdaiIgnoreTemplate = `# .stdaiignore: gitignore 风格 glob，匹配的源文件不参与 sync
# 路径相对 .stdai/standards/，支持 doublestar (**) 与 ? * 通配
# 行首 # 为注释，空行忽略

# 草稿文件（示例）
# rules/draft-*.md
# **/wip-*.md

# 内部文档不发到任何 target
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

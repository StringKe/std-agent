package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newApplyHooksCmd() *cobra.Command {
	var target string
	cmd := &cobra.Command{
		Use:   "apply-hooks",
		Short: "把 stdagent-hooks.json 的 hooks 字段 merge 到目标配置",
		Long: `把 stdagent 生成的中间 hooks 文件 merge 到目标工具实际加载的配置文件。

stdagent 默认不直接覆盖 .claude/settings.json（避免破坏用户其他设置），
而是写中间文件 .claude/stdagent-hooks.json。本命令负责把这个中间文件的
hooks 字段写回 .claude/settings.json，保留其他字段。

支持 target:
  claude-code   merge 到 .claude/settings.json (默认)
  codex         merge 到 .codex/config.toml [hooks] (UNKNOWN schema，仅 placeholder)`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runApplyHooks(cmd, target)
		},
	}
	cmd.Flags().StringVar(&target, "target", "claude-code", "目标工具 (claude-code | codex)")
	return cmd
}

func runApplyHooks(cmd *cobra.Command, target string) error {
	_, root := resolveConfigPath()
	switch target {
	case "claude-code":
		return applyClaudeHooks(cmd, root)
	case "codex":
		return applyCodexHooks(cmd, root)
	default:
		return fmt.Errorf("unsupported target %q (expected claude-code | codex)", target)
	}
}

func applyClaudeHooks(cmd *cobra.Command, root string) error {
	hooksPath := filepath.Join(root, ".claude/stdagent-hooks.json")
	settingsPath := filepath.Join(root, ".claude/settings.json")

	hooksData, err := os.ReadFile(hooksPath) //nolint:gosec
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s not found; run `stdagent sync` first", hooksPath)
		}
		return err
	}
	var hooksWrapper struct {
		Hooks json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(hooksData, &hooksWrapper); err != nil {
		return fmt.Errorf("parse %s: %w", hooksPath, err)
	}

	settings := map[string]json.RawMessage{}
	if existing, rerr := os.ReadFile(settingsPath); rerr == nil { //nolint:gosec
		if jerr := json.Unmarshal(existing, &settings); jerr != nil {
			return fmt.Errorf("parse existing settings.json: %w", jerr)
		}
	} else if !errors.Is(rerr, fs.ErrNotExist) {
		return rerr
	}

	settings["hooks"] = hooksWrapper.Hooks

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o750); err != nil {
		return err
	}
	tmp := settingsPath + ".tmp"
	if err := os.WriteFile(tmp, out, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tmp, settingsPath); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	cmd.Printf("Merged hooks into %s\n", settingsPath)
	return nil
}

func applyCodexHooks(cmd *cobra.Command, root string) error {
	hooksPath := filepath.Join(root, ".codex/stdagent-hooks.json")
	if _, err := os.Stat(hooksPath); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("%s not found; run `stdagent sync` first", hooksPath)
		}
		return err
	}
	cmd.Printf("Codex hooks 实际 schema 在公开文档中部分 UNKNOWN。\n")
	cmd.Printf("stdagent 已生成 %s 作为镜像参考，请按 codex 官方文档手动\n", hooksPath)
	cmd.Printf("把 hooks 字段写到 .codex/config.toml 的 [hooks] 表。\n")
	cmd.Printf("（v1.5 计划自动适配 codex schema 后由本命令直接 merge）\n")
	return nil
}

package runner

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"std-ai/internal/budget"
	"std-ai/internal/config"
	"std-ai/internal/parser"
	"std-ai/internal/source"
	"std-ai/internal/state"
	"std-ai/internal/transformer"
	"std-ai/internal/transformer/transformerutil"
	"std-ai/internal/writer"
)

// Options 控制 sync 行为
type Options struct {
	ProjectRoot string
	ConfigPath  string
	DryRun      bool
	NoPull      bool
	NoBackup    bool
	NoPrune     bool
	Strict      bool
	Targets     []string
	Version     string
}

// Result 是 sync 报告
type Result struct {
	Plans       []*writer.Plan
	Written     int
	Skipped     int
	Pruned      int
	PrunedPaths []string
	BackupDir   string
	SourceFiles int
	Docs        int
	Warnings    []string
}

// Sync 执行完整同步流
func Sync(opts Options) (*Result, error) {
	transformerutil.SetVersion(opts.Version)

	cfg, err := config.Load(opts.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if opts.DryRun {
		cfg.DryRun = true
	}

	res := &Result{}

	// 1. pull (如启用)
	if !opts.NoPull && cfg.AutoPull {
		for _, name := range sortedKeys(cfg.Sources) {
			src := cfg.Sources[name]
			if !src.Enabled {
				continue
			}
			cacheDir := filepath.Join(opts.ProjectRoot, ".stdai/cache", name)
			g := &source.Git{
				NameValue: name,
				URL:       src.URL,
				Branch:    src.Branch,
				CacheDir:  cacheDir,
				Paths:     src.Paths,
				Auth:      src.Auth,
				TokenEnv:  src.TokenEnv,
			}
			if err := g.Pull(); err != nil {
				if opts.Strict {
					return nil, fmt.Errorf("pull %s: %w", name, err)
				}
				res.Warnings = append(res.Warnings, fmt.Sprintf("pull %s: %v", name, err))
			}
		}
	}

	// 2. collect docs (local + sources)
	var allFiles []source.File

	localRoot := filepath.Join(opts.ProjectRoot, ".stdai/standards")
	localFiles, err := source.NewLocal(localRoot).Files()
	if err != nil {
		return nil, fmt.Errorf("read local: %w", err)
	}
	allFiles = append(allFiles, localFiles...)

	for _, name := range sortedKeys(cfg.Sources) {
		src := cfg.Sources[name]
		if !src.Enabled {
			continue
		}
		g := &source.Git{
			NameValue: name,
			CacheDir:  filepath.Join(opts.ProjectRoot, ".stdai/cache", name),
			Paths:     src.Paths,
		}
		files, err := g.Files()
		if err != nil {
			if opts.Strict {
				return nil, err
			}
			res.Warnings = append(res.Warnings, fmt.Sprintf("source %s: %v", name, err))
			continue
		}
		allFiles = append(allFiles, files...)
	}
	// 2.5 .stdaiignore 过滤（gitignore 风格 glob，支持 doublestar `**`）
	ignorePath := filepath.Join(opts.ProjectRoot, ".stdaiignore")
	ignore, ierr := source.LoadIgnoreFile(ignorePath)
	if ierr != nil {
		if opts.Strict {
			return nil, fmt.Errorf("load .stdaiignore: %w", ierr)
		}
		res.Warnings = append(res.Warnings, fmt.Sprintf(".stdaiignore: %v", ierr))
	} else if len(ignore.Patterns()) > 0 {
		filtered := make([]source.File, 0, len(allFiles))
		ignored := 0
		for _, f := range allFiles {
			if ignore.Match(f.Path) {
				ignored++
				continue
			}
			filtered = append(filtered, f)
		}
		allFiles = filtered
		if ignored > 0 {
			res.Warnings = append(res.Warnings, fmt.Sprintf("[ignore] skipped %d files via .stdaiignore", ignored))
		}
	}

	res.SourceFiles = len(allFiles)

	// 3. parse
	//   - 仅 .md / .markdown 文件参与 parse
	//   - skills/<name>/SKILL.md 是 SKILL package 主文件
	//   - skills/<name>/<subdir>/*.md（references / templates 等）是辅助文件，
	//     不 parse 为 Document，由 collectSkillPackageFiles 关联到主 SKILL Document
	docs := make([]*parser.Document, 0, len(allFiles))
	for _, f := range allFiles {
		if !isMarkdownPath(f.Path) {
			continue
		}
		if isSkillSubdirMarkdown(f.Path) {
			continue
		}
		d, err := parser.Parse(f.Path, f.Raw)
		if err != nil {
			if opts.Strict {
				return nil, err
			}
			res.Warnings = append(res.Warnings, err.Error())
			continue
		}
		docs = append(docs, d)
	}
	res.Docs = len(docs)

	// 3.1 收集 SKILL package 辅助文件（scripts/ references/ assets/ 等）
	collectSkillPackageFiles(docs, allFiles)

	// 3.2 budget 检查：每个 doc body 字节 + 全部 rules 累加 vs target 上限
	for _, d := range docs {
		for _, msg := range budget.CheckDocument(d) {
			res.Warnings = append(res.Warnings, "[budget] "+msg)
		}
	}
	for _, msg := range budget.CheckTotalRules(docs) {
		res.Warnings = append(res.Warnings, "[budget] "+msg)
	}

	// 3.5. load mcp.json (optional)
	mcpPath := filepath.Join(opts.ProjectRoot, ".stdai/standards/mcp.json")
	if data, rerr := os.ReadFile(mcpPath); rerr == nil { //nolint:gosec
		var mcp config.MCPConfig
		if jerr := json.Unmarshal(data, &mcp); jerr != nil {
			if opts.Strict {
				return nil, fmt.Errorf("parse mcp.json: %w", jerr)
			}
			res.Warnings = append(res.Warnings, fmt.Sprintf("mcp.json: %v", jerr))
		} else if len(mcp.Servers) > 0 {
			cfg.MCP = &mcp
		}
	} else if !errors.Is(rerr, fs.ErrNotExist) {
		if opts.Strict {
			return nil, fmt.Errorf("read mcp.json: %w", rerr)
		}
		res.Warnings = append(res.Warnings, fmt.Sprintf("mcp.json: %v", rerr))
	}

	// 4. plan + apply per target
	enabledTargets := opts.Targets
	if len(enabledTargets) == 0 {
		for _, name := range sortedKeys(cfg.Targets) {
			if cfg.Targets[name].Enabled {
				enabledTargets = append(enabledTargets, name)
			}
		}
	} else {
		sort.Strings(enabledTargets)
	}

	st, _ := state.Load(filepath.Join(opts.ProjectRoot, state.StateFile))
	if st.Targets == nil {
		st.Targets = map[string]state.Target{}
	}

	// 防 race / 误写：列出 git submodule 路径，sync 拒绝写入跨 submodule 边界。
	// 用户可能用 nested/<submodule>/root.md 误把 submodule 顶级 CLAUDE.md 当本仓嵌套；
	// runner 检测后 skip 这些 op + 报 warning。
	submodulePaths := listSubmodulePaths(opts.ProjectRoot)

	w := writer.NewWriter(opts.ProjectRoot, cfg.DryRun)
	bk := writer.NewBackup(filepath.Join(opts.ProjectRoot, ".stdai/backups"), cfg.BackupKeep)

	for _, name := range enabledTargets {
		tr, ok := transformer.Get(name)
		if !ok {
			if opts.Strict {
				return nil, fmt.Errorf("unknown target %q", name)
			}
			res.Warnings = append(res.Warnings, fmt.Sprintf("unknown target %q", name))
			continue
		}
		plan, err := tr.Plan(docs, cfg)
		if err != nil {
			if opts.Strict {
				return nil, fmt.Errorf("transform %s: %w", name, err)
			}
			res.Warnings = append(res.Warnings, fmt.Sprintf("transform %s: %v", name, err))
			continue
		}
		res.Plans = append(res.Plans, plan)

		// submodule 边界保护：拒绝写入到 submodule 内（防止误把 submodule 顶级 CLAUDE.md
		// 当作本仓嵌套从而覆盖 submodule 内容）。
		for i := range plan.Files {
			op := &plan.Files[i]
			if op.Skip {
				continue
			}
			if sp := submoduleContaining(op.Path, submodulePaths); sp != "" {
				op.Skip = true
				op.Reason = fmt.Sprintf("WARN: refused write into submodule %q (run stdagent inside that submodule instead)", sp)
				res.Warnings = append(res.Warnings, fmt.Sprintf("%s/%s: %s", name, op.Path, op.Reason))
			}
		}

		// 根文件体积检查（CLAUDE.md / AGENTS.md / GEMINI.md / 等）
		for _, f := range plan.Files {
			if !f.IsRoot {
				continue
			}
			for _, msg := range budget.CheckRootFile(name, f.Path, len(f.Content)) {
				res.Warnings = append(res.Warnings, "[budget] "+msg)
			}
		}

		// 收集 transformer 标记的 WARN reason（如 copilot/opencode SkillFiles 被忽略）
		for _, f := range plan.Files {
			if f.Skip || f.Reason == "" {
				continue
			}
			if strings.HasPrefix(f.Reason, "WARN") {
				res.Warnings = append(res.Warnings, fmt.Sprintf("%s/%s: %s", name, f.Path, f.Reason))
			}
		}

		if !opts.NoBackup && cfg.Backup && !cfg.DryRun {
			paths := make([]string, 0, len(plan.Files))
			for _, f := range plan.Files {
				paths = append(paths, f.Path)
			}
			dir, berr := bk.Snapshot(opts.ProjectRoot, paths)
			if berr == nil && dir != "" {
				res.BackupDir = dir
			}
		}

		written, skipped, aerr := w.Apply(plan)
		if aerr != nil {
			return nil, fmt.Errorf("apply %s: %w", name, aerr)
		}
		res.Written += written
		res.Skipped += skipped

		// current = 本次 plan 中"应当存在于磁盘"的 path 集合：
		// 包含未 skip 的（写入）+ Reason=="unchanged" 的（内容一致）。
		// 排除 transformer / submodule 标的 WARN skip（这些 path 不在我们管理范围）。
		current := make(map[string]struct{}, len(plan.Files))
		for _, f := range plan.Files {
			if !f.Skip || f.Reason == "unchanged" {
				current[f.Path] = struct{}{}
			}
		}

		// 孤儿 = 上次 state 记录、本次 current 不再包含。submodule 边界保护：
		// 落在 submodule 内的 path 永远不主动删（即使是我们之前写过的，谨慎为上）。
		var orphans []string
		if prev, ok := st.Targets[name]; ok {
			for p := range prev.Outputs {
				if _, still := current[p]; still {
					continue
				}
				if sp := submoduleContaining(p, submodulePaths); sp != "" {
					res.Warnings = append(res.Warnings, fmt.Sprintf("%s/%s: orphan in submodule %q skipped", name, p, sp))
					continue
				}
				orphans = append(orphans, p)
			}
		}
		sort.Strings(orphans)

		if !opts.NoPrune && len(orphans) > 0 {
			if cfg.DryRun {
				res.Pruned += len(orphans)
				res.PrunedPaths = append(res.PrunedPaths, prefixed(name, orphans)...)
			} else {
				absOrphans := make([]string, 0, len(orphans))
				for _, p := range orphans {
					full := filepath.Join(opts.ProjectRoot, p)
					if err := os.Remove(full); err != nil && !errors.Is(err, fs.ErrNotExist) {
						res.Warnings = append(res.Warnings, fmt.Sprintf("prune %s: %v", p, err))
						continue
					}
					absOrphans = append(absOrphans, full)
					res.Pruned++
					res.PrunedPaths = append(res.PrunedPaths, fmt.Sprintf("%s/%s", name, p))
				}
				cleanEmptyDirs(absOrphans)
			}
		}

		// state.Outputs 记录本次 current（含 unchanged），确保下次 sync 能识别孤儿。
		// 前版本只记非 skip 的 path，unchanged 文件会从 state 丢失追踪，prune 永远扫不到。
		out := map[string]string{}
		for _, f := range plan.Files {
			if f.Skip && f.Reason != "unchanged" {
				continue
			}
			out[f.Path] = writer.Checksum(f.Content)
		}
		st.Targets[name] = state.Target{
			LastSync: time.Now().UTC(),
			Outputs:  out,
		}
	}

	st.Version = "1.0"
	st.LastSync = time.Now().UTC()
	if !cfg.DryRun {
		_ = state.Save(filepath.Join(opts.ProjectRoot, state.StateFile), st)
	}

	return res, nil
}

// collectSkillPackageFiles 把每个 type=skills 的 Document 关联其 skill 目录下的辅助文件
//
// SKILL.md path 形如 "skills/code-review/SKILL.md"；同目录下其他文件（scripts/lint.sh、
// references/checklist.md、assets/template.md 等）作为 SkillFile 附到 Document.SkillFiles，
// path 字段是相对 skill 目录的相对路径。
//
// 仅扫已 parse 的 source.File（即 .md / .markdown）。非 markdown 辅助文件需要 source 层
// 扩展支持（v1.3 待办）。
func collectSkillPackageFiles(docs []*parser.Document, allFiles []source.File) {
	// 建立 skill 根目录（"skills/<name>/"）-> Document 的索引
	skillRoots := map[string]*parser.Document{}
	for _, d := range docs {
		if d.Type != parser.TypeSkills {
			continue
		}
		// SKILL.md 路径必须以 SKILL.md 结尾才视为 package 主文件
		dir, base := splitSkillPath(d.Path)
		if base != "SKILL.md" {
			continue
		}
		skillRoots[dir+"/"] = d
	}
	if len(skillRoots) == 0 {
		return
	}

	// 扫所有 source.File，按 skill 根目录前缀归类
	for _, f := range allFiles {
		for root, doc := range skillRoots {
			if !strings.HasPrefix(f.Path, root) {
				continue
			}
			rel := strings.TrimPrefix(f.Path, root)
			// 跳过 SKILL.md 本身（由 parser 已处理）
			if rel == "SKILL.md" {
				continue
			}
			doc.SkillFiles = append(doc.SkillFiles, parser.SkillFile{
				Path: rel,
				Raw:  f.Raw,
			})
			break
		}
	}
}

// splitSkillPath 把 "skills/code-review/SKILL.md" 拆为 ("skills/code-review", "SKILL.md")
func splitSkillPath(p string) (string, string) {
	idx := strings.LastIndex(p, "/")
	if idx < 0 {
		return "", p
	}
	return p[:idx], p[idx+1:]
}

// isMarkdownPath 判断 path 后缀是否为 markdown
func isMarkdownPath(p string) bool {
	lower := strings.ToLower(p)
	return strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown")
}

// isSkillSubdirMarkdown 判断是否为 SKILL package 子目录里的 markdown 辅助文件
//
//	"skills/code-review/SKILL.md"               -> false（顶层 SKILL.md）
//	"skills/code-review/references/check.md"    -> true（子目录辅助）
//	"skills/code-review/scripts/setup.md"       -> true
//	"rules/style.md"                            -> false（非 skills/ 子树）
func isSkillSubdirMarkdown(p string) bool {
	if !strings.HasPrefix(p, "skills/") {
		return false
	}
	// skills/<n>/SKILL.md 共 3 段；子目录辅助路径段数 >= 4
	return strings.Count(p, "/") >= 3
}

// listSubmodulePaths 跑 `git -C <root> submodule status`，返回 submodule 相对路径列表。
// 非 git 仓库 / 没 submodule 返回 nil。
func listSubmodulePaths(root string) []string {
	cmd := exec.Command("git", "-C", root, "submodule", "status") //nolint:gosec // root 由 caller 控制
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var paths []string
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		// 形如：" <hash> <path> (<branch>)" 或 "-<hash> <path>"
		if len(fields) >= 2 {
			paths = append(paths, fields[1])
		}
	}
	return paths
}

// submoduleContaining 检查 filePath 是否落在某 submodule 内。返回匹配的 submodule 路径，
// 没匹配返回空串。filePath 是相对项目根的 forward-slash 路径。
func submoduleContaining(filePath string, submodulePaths []string) string {
	for _, sp := range submodulePaths {
		if filePath == sp || strings.HasPrefix(filePath, sp+"/") {
			return sp
		}
	}
	return ""
}

// cleanEmptyDirs 删除被 prune 文件的父目录链（若已空）。
//
// 与 internal/cli/clean.go 的同名函数语义一致。两份拷贝（runner / cli）暂不抽公共：
// 二者直接依赖关系会形成 cli -> runner 反向，writer 包是更合适的归宿，留待第三处用例时再抽。
func cleanEmptyDirs(paths []string) {
	dirs := map[string]bool{}
	for _, p := range paths {
		dir := filepath.Dir(p)
		for dir != "." && dir != "" {
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dirs[dir] = true
			dir = parent
		}
	}
	dirList := make([]string, 0, len(dirs))
	for d := range dirs {
		dirList = append(dirList, d)
	}
	sort.Slice(dirList, func(i, j int) bool {
		return len(dirList[i]) > len(dirList[j])
	})
	for _, d := range dirList {
		_ = os.Remove(d)
	}
}

// prefixed 给路径列表统一加 "<target>/" 前缀，仅 DryRun 汇报用
func prefixed(target string, paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = fmt.Sprintf("%s/%s", target, p)
	}
	return out
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

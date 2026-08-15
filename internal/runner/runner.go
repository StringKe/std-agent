package runner

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/StringKe/std-agent/internal/budget"
	"github.com/StringKe/std-agent/internal/config"
	"github.com/StringKe/std-agent/internal/parser"
	"github.com/StringKe/std-agent/internal/source"
	"github.com/StringKe/std-agent/internal/state"
	"github.com/StringKe/std-agent/internal/transformer"
	"github.com/StringKe/std-agent/internal/transformer/transformerutil"
	"github.com/StringKe/std-agent/internal/writer"
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
	for _, msg := range budget.CheckTotalSkills(docs) {
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

	// 所有 target 先完成 plan，再执行任何写入。这样可以统一共享文件并在写盘前
	// 拒绝不同 target 对同一路径产生的不兼容内容，避免输出取决于 target 顺序。
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
	}

	if err := transformer.CanonicalizeSharedAGENTS(res.Plans, docs, cfg); err != nil {
		return nil, fmt.Errorf("canonicalize shared AGENTS.md: %w", err)
	}

	// 防 race / 误写：列出 git submodule 路径，sync 拒绝写入跨 submodule 边界。
	// 用户可能用 nested/<submodule>/root.md 误把 submodule 顶级 CLAUDE.md 当本仓嵌套；
	// runner 检测后 skip 这些 op + 报 warning。
	submodulePaths := listSubmodulePaths(opts.ProjectRoot)
	for _, plan := range res.Plans {
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
			}
		}
	}

	if err := validatePlanCollisions(res.Plans); err != nil {
		return nil, err
	}

	st, _ := state.Load(filepath.Join(opts.ProjectRoot, state.StateFile))
	if st.Targets == nil {
		st.Targets = map[string]state.Target{}
	}

	w := writer.NewWriter(opts.ProjectRoot, cfg.DryRun)
	bk := writer.NewBackup(filepath.Join(opts.ProjectRoot, ".stdai/backups"), cfg.BackupKeep)

	for _, plan := range res.Plans {
		name := plan.Target

		// 根文件体积检查（CLAUDE.md / AGENTS.md / GEMINI.md / 等）
		for _, f := range plan.Files {
			if !f.IsRoot {
				continue
			}
			for _, msg := range budget.CheckRootFile(name, f.Path, len(f.Content)) {
				res.Warnings = append(res.Warnings, "[budget] "+msg)
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

		// 统一收集 WARN reason（transformer 标记的、submodule 边界的、以及
		// Apply 阶段产生的如 JSONMerge 解析失败）。放在 Apply 之后才能拿全。
		for _, f := range plan.Files {
			if strings.HasPrefix(f.Reason, "WARN") {
				res.Warnings = append(res.Warnings, fmt.Sprintf("%s/%s: %s", name, f.Path, f.Reason))
			}
		}

		// current = 本次 plan 中"应当存在于磁盘"的 path 集合：
		// 包含未 skip 的（写入）+ Reason=="unchanged" 的（内容一致）。
		// 排除 transformer / submodule 标的 WARN skip（这些 path 不在我们管理范围）。
		current := make(map[string]struct{}, len(plan.Files))
		for _, f := range plan.Files {
			if !f.Skip || f.Reason == "unchanged" {
				current[f.Path] = struct{}{}
			}
			// JSONMerge 目标（crush.json / kilo.jsonc）是用户配置文件，即使本次
			// 因 JSONC 解析失败被 WARN skip 也永远不能进 prune 名单
			if f.JSONMerge {
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
		// v3 迁移：command-as-skill 从 .agents/skills/cmd-<n>/ 迁到 commands/<n>/，
		// 旧路径若从未进 state（或 state 重置）则 state 孤儿检测扫不到，Codex 仍会加载坏文件。
		// .codex/memories 废弃迁移同理（rules inline 到 AGENTS.md，详见函数注释）。
		legacy := legacyCmdPrefixedSkillOrphans(opts.ProjectRoot, plan)
		legacy = append(legacy, legacyCodexMemoriesOrphans(opts.ProjectRoot, plan)...)
		for _, p := range legacy {
			if _, still := current[p]; still {
				continue
			}
			if sp := submoduleContaining(p, submodulePaths); sp != "" {
				continue
			}
			orphans = append(orphans, p)
		}
		// state 孤儿与 legacy 扫描可能命中同一路径，去重避免 Pruned 重复计数
		sort.Strings(orphans)
		orphans = slices.Compact(orphans)

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

// legacyCmdPrefixedSkillOrphans 发现 v3 前 command-as-skill 遗留的 cmd-<name>/SKILL.md。
//
// 当 plan 含 .agents/skills/commands/<name>/SKILL.md 时，若磁盘仍存在
// .agents/skills/cmd-<name>/SKILL.md 则列入 prune（不在 state 也能清理）。
func legacyCmdPrefixedSkillOrphans(projectRoot string, plan *writer.Plan) []string {
	const (
		newPrefix   = ".agents/skills/commands/"
		legacySkill = ".agents/skills/cmd-"
	)
	var out []string
	for _, f := range plan.Files {
		if !strings.HasPrefix(f.Path, newPrefix) || !strings.HasSuffix(f.Path, "/SKILL.md") {
			continue
		}
		rest := strings.TrimPrefix(f.Path, newPrefix)
		name, _, ok := strings.Cut(rest, "/")
		if !ok || name == "" {
			continue
		}
		legacy := legacySkill + name + "/SKILL.md"
		if _, err := os.Stat(filepath.Join(projectRoot, legacy)); err != nil {
			continue
		}
		out = append(out, legacy)
	}
	return out
}

// legacyCodexMemoriesOrphans 发现旧版 codex transformer 遗留的 .codex/memories/ 产物。
//
// .codex/memories 已废弃：Codex 的 memories 是 ~/.codex/memories/ 用户级自动记忆系统
// （无项目级目录），项目级 .codex/ 是官方 Team Config 配置目录（config.toml /
// rules/*.rules execpolicy / skills/）。codex rules 现 inline 到 AGENTS.md，
// references / subagents 降级到 .agents/。旧路径若从未进 state（或 state 重置）则
// state 孤儿检测扫不到，这里主动扫描；只清理带 stdagent marker 的 .md 文件，
// 用户自行放置的文件不动。
func legacyCodexMemoriesOrphans(projectRoot string, plan *writer.Plan) []string {
	if plan.Target != "codex" {
		return nil
	}
	root := filepath.Join(projectRoot, ".codex", "memories")
	var out []string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr // 目录不存在 / 不可读直接视为无遗留
		}
		if !strings.HasSuffix(p, ".md") {
			return nil
		}
		content, rerr := os.ReadFile(p) //nolint:gosec // 路径来自 WalkDir 扫描 projectRoot 下固定目录
		if rerr != nil || !bytes.Contains(content, []byte("Generated by stdagent")) {
			return nil //nolint:nilerr // 读不了或无 marker 的文件不删，保守跳过
		}
		rel, rerr := filepath.Rel(projectRoot, p)
		if rerr != nil {
			return nil //nolint:nilerr // 相对化失败则跳过该文件，不中断扫描
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	return out
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

func validatePlanCollisions(plans []*writer.Plan) error {
	type plannedOutput struct {
		target    string
		content   []byte
		jsonMerge bool
	}
	seen := map[string]plannedOutput{}
	for _, plan := range plans {
		for _, op := range plan.Files {
			if op.Skip {
				continue
			}
			key := filepath.Clean(filepath.FromSlash(op.Path))
			prev, ok := seen[key]
			if !ok {
				seen[key] = plannedOutput{
					target:    plan.Target,
					content:   op.Content,
					jsonMerge: op.JSONMerge,
				}
				continue
			}
			if prev.jsonMerge == op.JSONMerge && bytes.Equal(prev.content, op.Content) {
				continue
			}
			return fmt.Errorf(
				"output collision at %q: targets %q and %q produce incompatible content",
				op.Path, prev.target, plan.Target,
			)
		}
	}
	return nil
}

func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

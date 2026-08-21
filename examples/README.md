# examples

本仓库源码只同步自己实际使用的 target（见根 `.stdai/config.toml`）。
完整 fan-out 与 gitignore 模式在这里用小型项目演示。

每个子目录都是独立项目根：进入后执行

```bash
go run ../../cmd/stdagent sync --no-pull --strict
```

| 目录 | `gitignore` | 展示 |
|---|---|---|
| [generated](generated/) | `generated`（默认） | 可重建产物全部忽略，只提交 `.stdai/` |
| [portable](portable/) | `portable` | 忽略专属目录，保留公约集合 `AGENTS.md` + `.agents/` |
| [off](off/) | `off` | 不改 `.gitignore`，生成物与源一起提交 |

三个例子都只开 `claude-code` 和 `codex`，源文件相同，差异只在 `gitignore` 与是否提交生成物。

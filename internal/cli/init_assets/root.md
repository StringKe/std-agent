# 项目名

用一句话说明项目的目标和主要使用者。

## 结构

- `<path>`：职责。

## 关键约束

- 只保留跨项目、无法从代码自然推断且违反会造成真实风险的约束。

## Done means

列出本项目完成变更时必须通过的最小验证命令。

AI 配置源位于 `.stdai/standards/`。修改后运行：

```bash
stdagent sync --strict
stdagent status
stdagent budget --rendered
```

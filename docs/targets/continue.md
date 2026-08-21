# Target: Continue.dev（已移除）

Continue 官方仓库已只读，且 README 写明 no longer actively maintained，Final 2.0.0 为最后发布。

- https://github.com/continuedev/continue/blob/main/README.md
- Cursor 收购后不消费 `.continue/`，无法迁移为现有 target

stdagent 自本次起从 `ValidTargets` 删除 `continue-dev`。旧 `config.toml` 若仍启用该名会校验失败。

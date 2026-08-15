# stdagent workflow

`.stdai/standards/` 是 source，生成的根文件和 target 目录是 rendered output。

## 首次迁移

1. 完整读取现有 AI 配置、嵌套指令、skills、commands 和 agents。
2. 盘点项目事实、硬约束、重复、冲突和唯一信息。
3. 将项目入口写入 `root.md`，其余内容按 rules、skills、commands、references、subagents 分类。
4. 删除角色化开场、过程性思考、重复规则和无信息示例；保留安全边界、协议字段、真实命令和成功标准。
5. 运行 `stdagent sync --strict`、`stdagent status`、`stdagent budget --rendered`。
6. 仅在新 source 已覆盖全部有效信息后，以可恢复方式处理旧产物。

## 日常维护

只编辑 `.stdai/standards/`，然后 sync 和验证。不要直接编辑带 stdagent marker 的文件。

`root.md` 只放高常驻价值的项目入口和硬约束，不手写自动 manifest。细节放专门 rule、skill 或 reference。

同一 git 仓库的目录级说明使用 `nested/<relative-path>/root.md`。Git submodule 独立维护自己的 `.stdai/`。

## 命令

| 命令 | 用途 |
|---|---|
| `stdagent sync --strict` | 生成配置并严格报告错误 |
| `stdagent status` | 检查 drift |
| `stdagent which <path>` | 查询适用 source |
| `stdagent budget --rendered` | 查看 source、root 和 sidecar 体积 |
| `stdagent clean` | 清理 state 管理的生成物 |

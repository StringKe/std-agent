# stdagent workflow

`.stdai/standards/` is the source; generated root files and target directories are rendered output.

## First-time migration

1. Fully read the existing AI configs, nested instructions, skills, commands, and agents.
2. Inventory project facts, hard constraints, duplicates, conflicts, and unique information.
3. Write the project entrypoint to `root.md`; classify everything else as rules, skills, commands, references, or subagents.
4. Delete role-play intros, process thinking, duplicate rules, and information-free examples; keep safety boundaries, protocol fields, real commands, and success criteria.
5. Run `stdagent sync --strict`, `stdagent status`, and `stdagent budget --rendered`.
6. Handle old artifacts in a recoverable way only after the new sources cover all valid information.

## Daily maintenance

Edit only `.stdai/standards/`, then sync and verify. Never edit files carrying a stdagent marker directly.

`root.md` keeps only the high resident-value project entrypoint and hard constraints; never hand-write the automatic manifest. Details go in dedicated rules, skills, or references.

Directory-level notes in the same git repo use `nested/<relative-path>/root.md`. Git submodules maintain their own `.stdai/` independently.

## Commands

| Command | Purpose |
|---|---|
| `stdagent sync --strict` | Generate config and report errors strictly |
| `stdagent status` | Check drift |
| `stdagent which <path>` | Query the applicable sources |
| `stdagent budget --rendered` | Show source, root, and sidecar sizes |
| `stdagent clean` | Clean state-managed artifacts |

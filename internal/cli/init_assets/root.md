# Project name

Describe the project's goal and main users in one sentence.

## Structure

- `<path>`: responsibility.

## Key constraints

- Keep only cross-project constraints that cannot be inferred from the code and whose violation causes real risk.

## Done means

List the minimal verification commands that must pass when changing this project.

AI config sources live in `.stdai/standards/`. After editing, run:

```bash
stdagent sync --strict
stdagent status
stdagent budget --rendered
```

---
name: wt
description: Manage isolated Git worktrees and feature branches using the wt CLI tool. Use when creating feature worktrees, switching contexts, syncing with upstream, inspecting worktree status, tearing down worktrees, or configuring .wt.toml.
---

# Git Worktree Management with `wt`

`wt` simplifies Git worktree workflows by provisioning isolated workspaces in `.worktrees/`, running post-creation hooks, and enforcing safety guardrails during teardown.

## Quick Start

### Create a Worktree
Provision a feature branch and worktree in `.worktrees/<branch>`:
```bash
wt new <branch>
```
Flags:
- `--from <ref>`: Branch from a specific ref instead of `origin/main` (or configured base branch).
- `--open`: Open the new worktree in `$EDITOR` or configured editor.

### Sync Worktree
Rebase current worktree onto latest remote base branch:
```bash
wt sync
```

### List Worktrees & Disk Usage
```bash
wt list    # Show active worktrees, branches, status, and age
wt size    # Report disk usage of all worktrees
```

### Refresh Remote Base Branch
Fetch remotes and update the local base branch pointer:
```bash
wt refresh
```

### Teardown a Worktree
Remove worktree and delete associated branch safely:
```bash
wt done          # Inside a worktree (uses current branch)
wt done <branch> # From main worktree
```
*Safety features:* Prompts to stash, push, force delete, or abort if working tree is dirty or branch is unmerged.

## Configuration (`.wt.toml`)

Optional configuration file placed at repository root:

```toml
[hooks]
post_create = "./scripts/setup.sh" # Custom post-creation script

[editor]
command = "code"                   # Editor binary for `wt new --open`

[sync]
base_branch = "main"               # Default base branch (default: auto-detected remote HEAD)
remote = "origin"                  # Remote name (default: "origin")
```

### Post-Create Hook Priority
If `hooks.post_create` is not configured, `wt` auto-detects setup tasks based on repo lockfiles:
1. `pnpm-lock.yaml` (`pnpm install`)
2. `bun.lockb` / `bun.lock` (`bun install`)
3. `yarn.lock` (`yarn install`)
4. `uv.lock` (`uv sync`)
5. `deno.json` / `deno.jsonc` (`deno install`)
6. `package.json` (`npm install`)
7. `go.mod` (`go mod download`)
8. `requirements.txt` (`pip install -r requirements.txt`)
9. `pyproject.toml` (`pip install -e .`)

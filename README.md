# wt

`wt` abstracts away the verbosity of standard `git worktree` commands, provides safety guardrails for branch management, and automates common workspace setup and teardown tasks.

## Installation

### Homebrew (macOS & Linux)

```bash
brew install servusdei2018/tap/wt
```

### Building from Source

```
git clone https://github.com/servusdei2018/wt
cd wt && make

# Place the binary where you keep your executables
# Ensure this directory is in your $PATH
cp bin/wt ~/.local/bin
```

Shell completions and man pages are available via the hidden `completion` and `man` subcommands added automatically:

```
wt completion bash > ~/.local/share/bash-completion/completions/wt
wt man > ~/.local/share/man/man1/wt.1
```

## Quick Start

```
$ wt help

  wt manages Git worktrees.                                                                                             
         
  USAGE  
                            
    wt [command] [--flags]                     
            
  COMMANDS  
            
    completion [command]      Generate the autocompletion script for the specified shell
    dedupe                    Retroactively deduplicate heavy build dependencies across worktrees
    done [branch]             Tear down a worktree and delete its branch
    help [command] [--flags]  Help about any command
    list                      List all worktrees with status and age
    new <branch> [--flags]    Create a new worktree for a feature branch
    refresh                   Fetch remotes and update the local base branch pointer
    size                      Report disk usage of all worktrees
    sync                      Rebase the current worktree onto the latest remote base branch
         
  FLAGS  
         
    -h --help                 Help for wt
    -v --version              Version for wt
```

### Automated Copy-On-Write Resolution

Heavy dependency directories (`node_modules`, `.venv`, `target/`, `vendor/`, etc.) can consume tens of gigabytes across multiple worktrees. `wt` solves this using OS-native Copy-On-Write (reflink) file cloning:

- **Instant Worktree Creation**: When running `wt new`, `wt` automatically pre-seeds heavy dependency directories from existing worktrees or the repository root using zero-copy reflinks before post-creation hooks run.
- **Retroactive Deduplication**: Run `wt dedupe` to scan existing worktrees and deduplicate duplicate dependency files using OS-native reflinks (`FICLONE` ioctl on Linux Btrfs/XFS/ZFS, APFS `clonefile` on macOS).
- **Dry Run**: Pass `--dry-run` to preview files and byte savings without altering disk state:
  ```bash
  wt dedupe --dry-run
  ```

## Configuration

wt works with zero configuration. An optional `.wt.toml` file at the repository root allows overriding defaults:

```toml
[hooks]
# Custom script to run after creating a new worktree.
# When set, this takes priority over auto-detected hooks.
post_create = "./scripts/setup.sh"

[editor]
# Override $EDITOR for 'wt new --open'.
command = "code"

[sync]
# Override the auto-detected base branch name.
base_branch = "main"
# Override the remote name (default: "origin").
remote = "origin"

[storage]
# Seed heavy dependency directories via CoW reflink on 'wt new' (default: true).
dedupe_on_create = true
# Fail if CoW reflink is unsupported on the filesystem (default: false).
require_reflink = false
# List of heavy dependency directory names to target.
heavy_dirs = ["node_modules", "vendor", ".venv", "venv", "__pycache__", ".gradle", "target", ".build", ".tox"]

[seed]
# Copy local config & secret files from repo root into new worktrees (default: [".env", ".env.local"]).
# Text files support Go template interpolation (e.g. {{ .Branch }}, {{ .Port }}).
include = [".env", ".env.local", "config/secrets.json"]
```

### Seed File Copying & Environment Secrets

When creating a new worktree with `wt new`, `wt` automatically copies uncommitted local configuration files (`.env`, `.env.local`, etc.) from the main repository root into the new worktree before post-creation hooks run.

Text configuration files support Go template interpolation for dynamic local environments:
- `{{ .Branch }}` — Feature branch name (e.g. `feature/login`)
- `{{ .BranchSlug }}` — Sanitized web-safe branch slug (e.g. `feature-login`)
- `{{ .Port }}` — Deterministically assigned unique port (`3000–8999`)
- `{{ .WorktreePath }}` — Absolute path of the new worktree
- `{{ .WorktreeName }}` — Directory basename of the worktree
- `{{ .RepoRoot }}` — Path to the main repository root

### Post-Create Hook Detection Priority

When `hooks.post_create` is not set, wt detects the environment automatically based on the following priority:

1. `pnpm-lock.yaml` — runs `pnpm install`
2. `bun.lockb` or `bun.lock` — runs `bun install`
3. `yarn.lock` — runs `yarn install`
4. `uv.lock` — runs `uv sync`
5. `deno.json` or `deno.jsonc` — runs `deno install`
6. `package.json` — runs `npm install`
7. `go.mod` — runs `go mod download`
8. `requirements.txt` — runs `pip install -r requirements.txt`
9. `pyproject.toml` — runs `pip install -e .`

### Base Branch Auto-Detection

When `sync.base_branch` is not set in `.wt.toml`, wt queries `git remote show
origin` to determine the remote HEAD branch. This handles repositories using
`main`, `master`, or any custom trunk name. Falls back to `main` if the remote is unreachable.

## Agent Skill

`wt` includes a [`SKILL.md`](SKILL.md) for AI coding agents. You may install it like so:

```
$ npx skills add servusdei2018/wt
```

## License

`wt` is distributed under the MIT License. See [LICENSE](LICENSE).


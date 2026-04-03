---
name: hush
description: |
  Manage private AI agent rules across projects and devices.
  Sync AGENTS.override.md files via a personal Git repo.
triggers:
  - hush
  - /hush
  - private rules
  - agent rules
  - sync rules
  - hush sync
  - hush edit
  - hush status
invocable: true
argument-hint: "[command] [args...]"
---

# /hush — Private Agent Rules Manager

Sync private `AGENTS.override.md` files across projects and devices via a personal Git repo.

## Agent Invariants

1. **Always check `hush status`** before suggesting sync or push operations
2. **Use `--json` flag** when parsing output programmatically
3. **Check breadcrumbs** in responses for suggested next actions

## Quick Reference

| Command | Purpose |
|---------|---------|
| `hush init <repo-url>` | Initialize with private Git repo |
| `hush link [--id name]` | Register current project |
| `hush sync [project]` | Pull + distribute override files |
| `hush push [project]` | Push local changes to repo |
| `hush edit [project]` | Open private rules in editor |
| `hush edit --public` | Open public AGENTS.md |
| `hush status` | Show all project sync status |
| `hush doctor` | Diagnose configuration issues |
| `hush hook install` | Ensure .gitignore is set up |

## How It Works

Each project gets three files:
- `AGENTS.md` (public, checked in) — ends with `@AGENTS.override.md`
- `CLAUDE.md` (public, checked in) — one line: `@AGENTS.md`
- `AGENTS.override.md` (private, gitignored) — synced by hush

Agent compatibility:
- **Claude Code**: CLAUDE.md → @AGENTS.md → @AGENTS.override.md
- **Codex**: reads AGENTS.override.md natively
- **OpenCode/KiloCode**: AGENTS.md → @AGENTS.override.md

## Common Workflows

### First-Time Setup
```bash
hush init git@github.com:user/my-rules.git
cd /path/to/project
hush link
hush edit
```

### Daily Use
```bash
hush sync          # Pull latest private rules
hush edit          # Edit private rules
hush push          # Push changes
```

### New Device
```bash
hush init git@github.com:user/my-rules.git
hush sync          # All projects restored
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | JSON envelope output |
| `--quiet` | Minimal output |
| `--agent` | Agent-optimized output |
| `--verbose` | Show debug details |
| `--dry-run` | Preview without writing |

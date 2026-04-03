# hush — Dotfiles for AI Agent Rules

Go CLI tool that syncs private `AGENTS.override.md` files across projects and devices via a personal Git repo.

## Dev & Test

```bash
make build        # Build → bin/hush
make test         # Run all tests
make test-unit    # Unit tests only
make lint         # golangci-lint
make vet          # go vet
make check        # Full CI gate (fmt + vet + lint + test)
```

## Architecture

- `cmd/hush/` — Cobra command definitions (init, sync, push, link, edit, hook, status, doctor)
- `internal/config/` — hush.yaml schema, ~/.hush/ path helpers
- `internal/project/` — Git remote detection, project ID generation, CWD resolution
- `internal/gitops/` — Git clone/pull/push wrappers, URL normalization
- `internal/sync/` — Core sync engine + scaffold (ensure AGENTS.md, CLAUDE.md, .gitignore)
- `internal/doctor/` — Diagnostic checks
- `internal/output/` — Response + Breadcrumbs output (styled/json/quiet/agent modes)
- `internal/skills/` — Embedded SKILL.md for ctx registry

## How It Works

Each managed project gets:
- `AGENTS.md` (public, checked in) — ends with `@AGENTS.override.md`
- `CLAUDE.md` (public, checked in) — one line: `@AGENTS.md`
- `AGENTS.override.md` (private, gitignored) — synced from private repo

Agent compatibility:
- Claude Code: CLAUDE.md → @AGENTS.md → @AGENTS.override.md (recursive @include)
- Codex: reads AGENTS.override.md natively as local override
- OpenCode/KiloCode: AGENTS.md → @AGENTS.override.md (@include)

@AGENTS.override.md

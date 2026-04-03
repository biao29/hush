# hush

Dotfiles for AI agent rules — sync private `AGENTS.override.md` across projects and devices via a personal Git repo.

## Install

**macOS / Linux:**

```bash
curl -fsSL https://raw.githubusercontent.com/biao29/hush/main/scripts/install.sh | bash
```

**Windows (PowerShell):**

```powershell
irm https://raw.githubusercontent.com/biao29/hush/main/scripts/install.ps1 | iex
```

**Go:**

```bash
go install github.com/getctx/hush/cmd/hush@latest
```

## How It Works

Each project gets three files:

```
project/
├── AGENTS.md              ← public rules (checked in), ends with @AGENTS.override.md
├── CLAUDE.md              ← one line: @AGENTS.md (checked in)
├── AGENTS.override.md     ← private rules (gitignored, managed by hush)
└── .gitignore             ← includes AGENTS.override.md
```

All major AI coding agents read `AGENTS.override.md` through their native mechanisms:

- **Claude Code**: `CLAUDE.md` → `@AGENTS.md` → `@AGENTS.override.md` (recursive @include)
- **Codex**: reads `AGENTS.override.md` natively as highest-priority local override
- **OpenCode / KiloCode**: `AGENTS.md` → `@AGENTS.override.md` (@include)

No hooks. No stripping. Physical isolation via `.gitignore`.

## Quick Start

```bash
# 1. Initialize with your private rules repo
hush init git@github.com:you/my-rules.git

# 2. Register a project
cd ~/projects/my-api
hush link

# 3. Edit private rules
hush edit

# 4. Push to your private repo
hush push

# 5. On another device — restore everything
hush init git@github.com:you/my-rules.git
hush sync
```

## Verify It Works

After setup, create a canary rule to confirm the @include chain is loaded by your agent:

```bash
# Write a test rule
cat > AGENTS.override.md << 'EOF'
When asked "what is the hush canary?", respond with:
"canary: pineapple-on-pizza-42"
EOF

# Start a new agent session in the same directory
claude   # or codex, opencode, etc.
```

Ask: **"what is the hush canary?"**

If the agent responds `canary: pineapple-on-pizza-42`, the full chain works:

```
CLAUDE.md → @AGENTS.md → @AGENTS.override.md → agent reads private rules ✓
```

Confirm the file is protected from commits:

```bash
git status                    # AGENTS.override.md should NOT appear
git add AGENTS.override.md    # should be rejected by .gitignore
```

## Commands

| Command | Purpose |
|---------|---------|
| `hush init <repo-url>` | Clone private rules repo |
| `hush link [--id name]` | Register current project |
| `hush sync [project]` | Pull + distribute override files |
| `hush push [project]` | Push local changes to repo |
| `hush edit [project]` | Open private rules in `$EDITOR` |
| `hush edit --public` | Edit public AGENTS.md |
| `hush status` | Show sync status for all projects |
| `hush doctor` | Diagnose configuration issues |
| `hush hook install` | Ensure .gitignore is set up |
| `hush skill install` | Install agent skill to standard locations |
| `hush version` | Print version |

All commands support `--json`, `--quiet`, `--agent`, `--dry-run`, `--verbose`.

## License

MIT

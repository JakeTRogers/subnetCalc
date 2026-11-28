# subnetCalc

CLI for IPv4/IPv6 subnet math from CIDR notation: network details, splitting a supernet into subnets, and an interactive Bubble Tea TUI for splitting and joining.

## Architecture

- CLI flow: `cmd/root.go` → `subnet.NewNetwork()` → optional `Network.Split()` → `formatter.New()` (table, JSON, or text output).
- TUI flow (`-i`): `tui/` wraps networks in a `SubnetNode` tree for interactive split/join, undo, and JSON export.
- `internal/ui` holds the lipgloss styles and prefix-depth color palette shared by the table formatter and the TUI.

## Conventions

- Go style: `.github/instructions/go.instructions.md`.
- Host counts are `*big.Int` end to end, because IPv6 counts overflow fixed-width integers. The TUI's `SubnetNode.Hosts()` clamps to `^uint(0)` for display only.
- Tests use the standard `testing` package (no testify), live beside the code, call `t.Parallel()`, and are named `TestFuncName_scenario`.
- Conventional Commits enforced by commitizen (`.cz.yaml`) via the pre-commit `commit-msg` hook.

## Gotchas

- The pre-commit coverage hook fails if total test coverage drops below 85%.
- The version string (`v`-prefixed) lives in `cmd/root.go`; CI (`verBumpChkr.yml`) fails a PR whose commits call for a release (`feat`, `fix`, `refactor`, `perf`, breaking) without a matching bump, or that bumps it when none do. It is managed by `cz bump` (`version_files` in `.cz.yaml`).
- `Network.Split()` returns an error rather than generating more than `subnet.MaxGeneratedSubnets` (1,000,000) subnets.

## Agent skills

### Issue tracker

Issues are tracked in GitHub Issues on `JakeTRogers/subnetCalc` via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Default triage labels: `needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `GLOSSARY.md` and `docs/adr/` at the repo root. See `docs/agents/domain.md`.

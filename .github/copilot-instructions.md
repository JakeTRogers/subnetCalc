# Copilot Instructions for subnetCalc

## Project Overview

subnetCalc is a Go 1.25 CLI tool for calculating IPv4/IPv6 subnet information from CIDR notation. It can split supernets into subnets and provides both a CLI interface (table/JSON output) and an interactive TUI built with Bubble Tea.

## Architecture

```text
main.go              → Entry point, calls cmd.Execute()
cmd/root.go          → Cobra command, CLI flags, output orchestration
logger/logger.go     → Zerolog wrapper with verbosity levels (-v to -vvvv)
subnet/network.go    → Core subnet calculations and Network type
formatter/           → Output formatters: JSON, table (lipgloss), text
internal/ui/         → Shared lipgloss styles (styles.go)
tui/                 → Bubble Tea TUI:
  model.go           → State management
  tree.go            → SubnetNode tree with embedded Network
  render.go          → Main render orchestration
  render_format.go   → IP range abbreviation helpers
  render_cells.go    → Cell span calculations
  keys.go            → Key bindings
```

**Data flow:** CIDR string → `subnet.NewNetwork()` → optional `Split()` → `formatter.New()` → output
**TUI flow:** CIDR → `tui.NewModel()` → `SubnetNode` tree → interactive split/join → optional JSON export

**Key types:**

- `subnet.Network` – holds CIDR, network/broadcast addresses, subnet mask, host count, optional `[]Subnets`
- `SubnetNode` (tui) – embeds `subnet.Network`, tree structure with parent/children for interactive split/join
- `formatter.Formatter` – interface with `FormatNetwork()` and `FormatSubnets()` implementations
- `ui.PrefixColors` – shared color palette for consistent prefix-depth styling

## Code Patterns

**Error handling:** Wrap errors with context using `fmt.Errorf("description: %w", err)`

**Logging:** Use package-level logger from `logger.GetLogger()`:

```go
log := logger.GetLogger()
log.Debug().Str("cidr", cidr).Msg("parsing network")
log.Err(err).Msg("failed to calculate")
```

**Logger integration:** CLI sets verbosity in `PersistentPreRun` via wrapper:

```go
PersistentPreRun: func(cmd *cobra.Command, args []string) {
    verboseCount, _ := cmd.Flags().GetCount("verbose")
    logger.SetLogLevel(verboseCount)
}
```

**Adding CLI flags:** Define in `init()`, use `rootCmd.Flags()` for command-specific, `PersistentFlags()` for global. Mutually exclusive flags use `MarkFlagsMutuallyExclusive("interactive", "json")`.

**Big.Int for IPv6:** Use `math/big` for host calculations to avoid overflow on large IPv6 subnets.

**Subnet safety:** `subnet.MaxGeneratedSubnets` (1,000,000) prevents OOM from very large splits.

**Shared styles:** Import styles from `internal/ui` for consistent formatting:

```go
import "github.com/JakeTRogers/subnetCalc/internal/ui"

color := ui.GetColorForPrefix(bits, initialPrefix)
style := ui.HeaderStyle.Render("Header")
```

## Development Commands

```bash
go test ./... -v                    # Run all tests
go test ./... -cover                # Run with coverage
go build -o subnetCalc .            # Build binary
./subnetCalc 10.0.0.0/8 -i          # Test TUI
./subnetCalc 192.168.0.0/24 -s 26   # Test subnet splitting
```

**Pre-commit hooks:** Run `golangci-lint`, `go test`, and commitizen checks. Install with `pre-commit install`.

## Testing Patterns

Tests are co-located with source (`*_test.go`). Test names follow `TestFuncName_scenario` pattern for consistency.

Key test patterns:

- Integration tests in `main_test.go` capture CLI output and verify formatting
- TUI tests in `tui/model_test.go` verify state transitions and tree operations
- Subnet calculation tests in `subnet/network_test.go` verify IPv4/IPv6 math

Example test structure:

```go
func TestNewNetwork_validIPv4(t *testing.T) {
    t.Parallel()
    n, err := subnet.NewNetwork("10.0.0.0/24")
    require.NoError(t, err)
    // assertions...
}
```

## IPv6 Considerations

- IPv6 is supported but limited: very large subnet splits are capped
- Broadcast addresses shown as last IP in range (IPv6 has no broadcast)
- TUI caps host display at `^uint(0)` for IPv6 with >32 host bits

## Commit Convention

Uses [Conventional Commits](https://www.conventionalcommits.org/) enforced by pre-commit hooks via commitizen. Version managed in `.cz.yaml` and auto-updated in `cmd/root.go`.

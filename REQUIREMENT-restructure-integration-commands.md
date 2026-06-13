# Feature: Restructure integration command hierarchy

## Overview
Change the `agent-hub integration` command structure from `<action> <runner>` ordering to `<runner> <action>` ordering.

## Old → New mapping

```
integration install opencode      → integration opencode install
integration status opencode       → integration opencode status
integration uninstall opencode    → integration opencode uninstall
integration enable opencode       → integration opencode enable
integration disable opencode      → integration opencode disable
integration status                → integration status  (unchanged)
```

## Implementation

In `cmd/agent-hub/main.go`, restructure `runIntegration()`:

1. `integration` with no args → show help (unchanged)
2. `integration --help` → show help (unchanged)
3. `integration status` → show all runners (unchanged)
4. `integration <runner>` where runner is "opencode" → dispatch to runner subcommand:
   - `integration opencode status` → show opencode status
   - `integration opencode install [--global]` → install plugin
   - `integration opencode uninstall [--global]` → remove plugin
   - `integration opencode enable [--global]` → enable plugin
   - `integration opencode disable [--global]` → disable plugin
5. Unknown runner → error
6. Unknown subcommand under known runner → error

### `--global` flag support
All of install, uninstall, enable, disable support `--global`. Without `--global`, operate on local `.opencode/plugins/`. With `--global`, operate on `$HOME/.config/opencode/plugins/`.

Refactor a shared helper:
```go
func integrationPluginsDir(global bool) (string, error) { ... }
```

### Help text updates
- `integration --help`: lists "status" and "opencode" as subcommands
- `integration opencode --help`: lists status/install/uninstall/enable/disable
- Each `integration opencode <subcommand> --help`: describes the command and --global flag

## Test changes (already applied, sealed)
- 17 integration SETUP.md files: args reordered from `[integration, <action>, opencode]` to `[integration, opencode, <action>]`
- 6 help test files restructured (old removed, new created under `integration-opencode-*`)
- 3 new global tests: uninstall/global, enable/global, disable/global
- Edge test: `integration opencode unknown` instead of `integration unknown-subcommand`
- Help assertions for top-level `integration` now expect "opencode" and "status" (not install/uninstall/etc.)

## Verification
```sh
doctest test -v ./agents/agent-hub/tests/integration
doctest test -v ./agents/agent-hub/tests/help
```

## Constraints
- Tests are sealed (staged) — must NOT be modified
- Also run existing tests to ensure no regressions: `doctest test -v ./agents/agent-hub/tests`
- `go build ./cmd/agent-hub` must succeed

# Scenario

**Feature**: CLI ask command with human checkpoint semantics

```
debug-with-user ask --title ... --option ... -> dialog.Ask -> JSON stdout or exit 1/2
```

## Preconditions

- `ask` subcommand is registered on the CLI.
- Dry-run mode is enabled for all leaves except `non-mac`.

## Steps

1. Start from `defaultAskArgs()` unless a leaf overrides flags.
2. Append dry-run env vars per scenario.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	if len(req.Args) == 0 {
		req.Args = defaultAskArgs()
	}
	return nil
}
```

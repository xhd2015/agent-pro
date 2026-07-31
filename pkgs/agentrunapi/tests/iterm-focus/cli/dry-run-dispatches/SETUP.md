# Scenario

**Feature**: CLI dry-run path does not focus

```
# Prefer library FocusSession with DryRun for deterministic inject (L2).
# This leaf locks CLI arg surface: RunFocus(["--dry-run", "<session-id>"]) must
# not return "unknown command" once wired; when store/env injection is thin,
# assert dry-run flag is accepted (help-adjacent) OR full resolve when
# implementer threads FocusOpts from CLI.

# Concrete L2 contract for this leaf:
# RunFocus with --dry-run and a session id returns without panic; either:
#   (a) nil error + stdout describing candidate / dry-run, or
#   (b) non-nil resolve error that is NOT "unknown" / "not implemented" shape
#       after full implement — until then RED is fine.
# Strong guarantee reused from library: dry-run never focuses (covered by
# library/single-match/dry-run). Here we lock CLI flag parsing + dispatch.
```

## Steps

1. CLIArgs: `--dry-run` + session id (and optional `--session-id` form is implementer choice; positional is primary).

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Session need not exist: we lock dispatch + dry-run flag acceptance.
	// Unknown session should error as resolve failure, not "unknown command".
	req.CLIArgs = []string{"--dry-run", "sess-cli-dry"}
	return nil
}
```

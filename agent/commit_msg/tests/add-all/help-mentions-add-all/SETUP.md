# Scenario

**Feature**: `-h` help text documents `--add-all`

```
gen-commit-msg -h
  -> exit 0
  -> help mentions --add-all
```

## Preconditions
- Help path builds/runs `cmd/gen-commit-msg` subprocess (`req.Help = true`).
- No git repo required for help.

## Steps
1. Set `req.Help = true`.
2. Run gen-commit-msg help.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Help = true
	req.AddAll = false
	req.Commit = false
	req.DryRun = false
	req.Operation = "help-mentions-add-all"
	return nil
}
```

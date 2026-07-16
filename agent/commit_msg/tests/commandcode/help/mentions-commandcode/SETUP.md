# Scenario

**Feature**: `-h` help text contains `commandcode`

```
gen-commit-msg -h
  -> exit 0
  -> help mentions commandcode (and opencode) under --agent-runner
```

## Preconditions
- Parent sets `req.Help = true` and builds the CLI binary.

## Steps
1. Inherit Help flag from parent.
2. Run gen-commit-msg help.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Help = true
	req.Operation = "help-mentions-commandcode"
	return nil
}
```

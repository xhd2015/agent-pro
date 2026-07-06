# Scenario

`--no-verify` without `--commit` must fail at flag validation before invoking the agent.

## Preconditions
- No git repository or fake-opencode mock is required; validation is pre-agent.

## Steps
1. Set `req.NoVerify = true` and leave `req.Commit` false.
2. Run gen-commit-msg with only `--no-verify`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.NoVerify = true
	req.Commit = false
	return nil
}
```
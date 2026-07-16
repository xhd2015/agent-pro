# Scenario

**Feature**: help documents commandcode as a supported agent runner

```
# help path
gen-commit-msg -h
  -> usage text lists agent-runner values including commandcode
```

## Preconditions
- `cmd/gen-commit-msg` binary built by parent commandcode Setup (`req.GenCommitMsgBin`).
- No git repo required for help.

## Steps
1. Set `req.Help = true` so Run uses the CLI subprocess with `-h`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Help = true
	req.Commit = false
	req.DryRun = false
	return nil
}
```

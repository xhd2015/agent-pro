# Scenario

**Feature**: dry-run plans binary unstage on stderr without mutating the index

```
# stage binary + text; dry-run does not unstage
staged: app.go + blob.bin -> gen-commit-msg --dry-run
  -> stderr: would: unstage … blob.bin
  -> index still has blob.bin staged
  -> mock N = 2 (count before unstage)
```

## Preconditions
- Repo stages one text file and one ELF-like binary.
- Pure dry-run: agent binary is non-existent.

## Steps
1. Initialize repo; stage `app.go` and `blob.bin`.
2. Set `req.DryRun = true` and a non-existent agent binary.
3. Run gen-commit-msg.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	binRel := StageDryRunRepoWithBinary(t, req)
	req.DryRun = true
	req.Commit = false
	req.AgentRunnerBinary = NonExistentAgentBinary(req)
	req.Operation = binRel // binary relative path for assert
	return nil
}
```

# Scenario

**Feature**: relative `--dir` resolves against process cwd

```
cwd=TempDir; --dir rel-ws
  -> meta.workspace = abs(TempDir/rel-ws)
```

## Steps

1. Leaves create a relative workspace directory under process cwd.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Relative-path branch; process cwd is req.TempDir (root runAgentRun cmd.Dir).
	t.Helper()
	if req.TempDir == "" {
		t.Fatal("TempDir must be set by root Setup before relative --dir leaves")
	}
	return nil
}
```

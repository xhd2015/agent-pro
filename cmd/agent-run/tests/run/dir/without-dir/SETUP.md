# Scenario

**Feature**: omitted `--dir` keeps today's Getwd workspace default

```
agent-run run --agent-runner fake-codex "hi"  # no --dir
  -> meta.workspace = process cwd (Getwd / harness TempDir)
```

## Steps

1. Leaves run without `--dir`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	// Omitted --dir branch: leaves must not append --dir.
	t.Helper()
	for _, a := range req.Args {
		if a == "--dir" || len(a) >= 6 && a[:6] == "--dir=" {
			t.Fatalf("without-dir grouping must not already include --dir; args=%v", req.Args)
		}
	}
	return nil
}
```

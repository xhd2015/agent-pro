# Scenario

**Feature**: session update --dir stores absolute path and preserves fields

```
session update --session-id ID --dir PATH
  -> OK session=ID dir=<abs>
  -> sessions.json dir=abs; other fields unchanged; updated_at newer
```

## Steps

1. Create workspace directory.
2. Args: session update --session-id --dir.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	abs, err := ensureUpdateWorkspaceDir(t, req)
	if err != nil {
		return err
	}
	req.Args = []string{
		"session", "update",
		"--session-id", sessionUpdateFixtureID,
		"--dir", abs,
	}
	return nil
}
```

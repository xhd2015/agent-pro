# Scenario

**Feature**: when grok session dir is missing, warn on stderr and fall back to scrollback capture

```
GROK_HOME has no matching session dir
  -> stderr warns about discovery failure
  -> end-of-run scrollback still emits fake TUI Response: hi
```

## Steps

1. Set empty `GROK_HOME` temp dir with no `sessions/` tree.
2. Run with respond fake TUI and prompt `hi`.
3. Assert stderr warning plus scrollback fallback on stdout/events.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.GrokHome = filepath.Join(req.TempDir, "empty-grok-home")
	if err := os.MkdirAll(req.GrokHome, 0755); err != nil {
		return err
	}
	req.SkipGrokSessionDir = true
	appendGrokHomeEnv(req)

	req.GrokTTYCommand = fakeTUIRespondHi()
	appendGrokTTYEnv(req)
	req.Args = append(req.Args, "hi")
	return nil
}
```
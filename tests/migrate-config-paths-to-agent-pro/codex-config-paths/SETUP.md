# Scenario

**Feature**: Codex config path function returns the correct default path.

## Preconditions

- `HOME` is set to the isolated temp directory.

## Steps

1. Set `HOME` to the isolated directory.
2. Call `codexconfig.DefaultConfigPath()` via Run.
3. Verify the returned path matches `$HOME/.codex/config.toml`.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Home = t.TempDir()
	req.TestCase = "codex-config"
	return nil
}
```

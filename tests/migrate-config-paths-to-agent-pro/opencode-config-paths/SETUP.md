# Scenario

**Feature**: Opencode config path function returns the correct global user path.

## Preconditions

- `HOME` is set to the isolated temp directory.

## Steps

1. Set `HOME` to the isolated directory.
2. Call `opencodeconfig.GlobalUserConfigPath()` via Run.
3. Verify the returned path matches `$HOME/.config/opencode/opencode.jsonc`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Home = t.TempDir()
	req.TestCase = "opencode-global-user"
	return nil
}
```

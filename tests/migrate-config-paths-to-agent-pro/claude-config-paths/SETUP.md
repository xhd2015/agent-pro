# Scenario

**Feature**: Claude config path functions return correct paths resolved from HOME.

## Preconditions

- `HOME` is set to the isolated temp directory.

## Steps

1. Set `HOME` to the isolated directory.
2. Call `SettingsPath()`, `JSONConfigPath()`, and `GlobalSkillsDir()` via Run.
3. Verify all returned paths.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Home = t.TempDir()
	req.TestCase = "claude-all"
	return nil
}
```

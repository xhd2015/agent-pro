# Scenario

**Feature**: Codex skill directory path functions return correct global dirs.

## Preconditions

- `HOME` is set to the isolated temp directory.

## Steps

1. Set `HOME` to the isolated directory.
2. Call `GlobalSkillDirs()` via Run.
3. Verify the returned paths.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Home = t.TempDir()
	req.TestCase = "codex-skills-global"
	return nil
}
```

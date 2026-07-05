# Scenario

**Feature**: Opencode local skill dirs return project-local paths.

## Preconditions

- `HOME` is set to the isolated temp directory.
- `req.ProjectDir` is set.

## Steps

1. Set `HOME` to the isolated directory.
2. Call `LocalSkillDirs(projectDir)` via Run.
3. Verify the returned paths.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Home = t.TempDir()
	req.TestCase = "opencode-skills-local"
	req.ProjectDir = t.TempDir()
	return nil
}
```

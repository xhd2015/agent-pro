# Scenario

**Feature**: full anti_patterns fixture corpus sanitizes or rejects as declared

```
# walk agent/commit_msg/testdata/anti_patterns/*.in
# each .want -> stdout equals formatted message
# each .want_err -> hard failure (no usable message)
```

## Preconditions
- All shared fixtures under `testdata/anti_patterns/` are present.
- Covers fixtures without dedicated named leaves (bold wrap, triple fence, unclosed tick, md label).

## Steps
1. Stage a single repo reused across fixture runs.
2. Placeholder mock so root `Run` can execute once.
3. Assert re-runs every fixture via `RunAntiPatternFixture`.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	StageRepoWithChange(t, req)
	// Placeholder: clean fixture so the mandatory Run path is well-formed.
	WriteMockAgentText(t, req, "sess_corpus_placeholder", ReadAntiPatternIn(t, "clean_json_unchanged"))
	req.Commit = false
	req.Operation = "sanitize_fixtures"
	return nil
}
```

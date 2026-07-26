# Scenario

**Feature**: bound resume with empty prior still injects only the new prompt

```
resumeID="grok-sess-abc" + prior=[] + "only"
  -> ResolveRunnerPrompt -> "only"
```

## Preconditions

- `ResumeID` is non-empty.
- `PriorEvents` is empty (no transcript rows yet, or none relevant).
- New prompt is `only`.

## Steps

1. Set resume id, empty prior, and new prompt.
2. Expect no regression: inject remains the trimmed new prompt only.

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ResumeID = "grok-sess-abc"
	req.NewPrompt = "only"
	req.PriorEvents = nil
	return nil
}
```

## Expected

- ScannedFiles > 0.
- At least one production file references `BuildFollowUpCommand`.
- That wire assigns `Model:` (not only ModelReasoningEffort) on FollowUpOpts.
- That wire assigns `ModelReasoningEffort:` (or `ModelReasoningEffort =`).

## Side Effects

- None (read-only scan).

## Errors

- None.

## Exit Code

N/A

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertNoError(t, err)
	if resp.ScannedFiles == 0 {
		t.Fatal("expected to scan pkgs/agentruncli .go files")
	}
	if !resp.BuildFollowUpRef {
		t.Fatal("pkgs/agentruncli must call BuildFollowUpCommand for ForceNew follow-up")
	}
	if !resp.ModelFieldFound {
		t.Fatal("ForceNew follow-up must assign FollowUpOpts.Model when plumbing model pass-through")
	}
	if !resp.EffortFieldFound {
		t.Fatal("ForceNew follow-up must assign FollowUpOpts.ModelReasoningEffort when plumbing effort pass-through")
	}
}
```

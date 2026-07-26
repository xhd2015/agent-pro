## Expected

- ScannedFiles > 0.
- Import of `pkgs/agentrunapi` present (P1 may already satisfy).
- Symbol `BuildFollowUpCommand` appears in `pkgs/agentruncli` production sources (new-terminal wire).

## Side Effects

- None (read-only).

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
	if !resp.ImportFound {
		t.Fatal("pkgs/agentruncli must import pkgs/agentrunapi")
	}
	if !resp.SymbolFound {
		t.Fatal("pkgs/agentruncli must call/reference BuildFollowUpCommand for new-terminal FollowUp")
	}
}
```

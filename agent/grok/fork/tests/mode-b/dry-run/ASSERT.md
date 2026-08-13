## Expected Output

```
Would fork grok session
  grok id:   019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa
  cwd:       <abs>
  command:   <grok-bin> --resume 019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa --fork-session
  terminal:  current
```

Output ends with a newline.

## Expected

- Exit 0.
- Locked plan; command uses grok-bin (basename `llm-mock-run-grok`).
- No OpenInNewTerminal.

## Side Effects

- None.

## Errors

- None.

## Exit Code

0

```go
import (
	"path/filepath"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	assertNoOpen(t, resp)
	assertNoForeground(t, resp)
	if filepath.Base(req.GrokBin) != "llm-mock-run-grok" {
		t.Fatalf("test grok bin basename: %q", req.GrokBin)
	}
	assertStdoutExact(t, resp.Stdout,
		"Would fork grok session",
		"  grok id:   "+fixtureSessionID,
		"  cwd:       "+req.Workspace,
		"  command:   "+req.GrokBin+" --resume "+fixtureSessionID+" --fork-session",
		"  terminal:  current",
	)
}
```

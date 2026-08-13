## Expected Output

```
Would open new iTerm2 window
  ancestor pid: 4242
  grok id:      019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa
  cwd:          <abs workspace>
  command:      <executable> --session-id 019f283b-aaaa-7aaa-aaaa-aaaaaaaaaaaa
```

Output ends with a newline.

## Expected

- Exit 0; `Main` nil error.
- Locked plan lines (cwd/exe interpolated).
- Command does **not** contain `grok --resume`.
- `OpenInNewTerminal` not called.

## Side Effects

- None (dry-run).

## Errors

- None.

## Exit Code

0

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = err
	assertMainOK(t, resp)
	assertNoOpen(t, resp)
	assertNoForeground(t, resp)
	assertNoANSI(t, resp.Stdout, "dry-run stdout")
	cmd := followUpSession(req.Executable, fixtureSessionID)
	assertStdoutExact(t, resp.Stdout,
		"Would open new iTerm2 window",
		"  ancestor pid: 4242",
		"  grok id:      "+fixtureSessionID,
		"  cwd:          "+req.Workspace,
		"  command:      "+cmd,
	)
	if strings.Contains(resp.Stdout, "grok --resume") {
		t.Fatalf("Mode A plan must not contain grok --resume:\n%s", resp.Stdout)
	}
}
```

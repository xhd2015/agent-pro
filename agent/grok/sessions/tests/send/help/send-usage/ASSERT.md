## Expected

- Help documents session source and send/open flags.

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	lower := strings.ToLower(resp.Stdout)
	for _, want := range []string{
		"--session-id", "--tab", "--tab-index", "--index",
		"--no-submit", "--focus", "--no-ctrl-u", "--open", "--no-agent-run", "--dir", "--dry-run",
		"--enter", "--up", "--ctrl-c", "--esc", "--text",
		"sendtext", "kool", "agent-run",
	} {
		if !strings.Contains(lower, strings.ToLower(want)) {
			t.Fatalf("send help must document %q:\n%s", want, resp.Stdout)
		}
	}
	assertNoSend(t, resp)
}
```

## Expected

- Ambiguity lists candidates with `snapshot` hint; never captures.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoContents(t, resp)
	assertErrorOutput(t, resp, `---
version: 3
---
multiple iTerm2 tabs host session `+req.SessionID+`

  \[0\] window 1 \("credit-pricing"\) tab 2 tty /dev/ttys148 session w0t2p0
  \[1\] window 3 \("worktrees"\) tab 1 tty /dev/ttys149 session w2t1p0

Specify one with:
  agent-pro grok session snapshot `+req.SessionID+` --index <0\|1>
`)
}
```

## Expected

- Ambiguity is fatal, reports every 0-based candidate, includes `--index`, and never focuses.

## Expected Output

```text
multiple iTerm2 tabs host session <id>

  [0] window 1 ("credit-pricing") tab 2 tty /dev/ttys148 session w0t2p0
  [1] window 3 ("worktrees") tab 1 tty /dev/ttys149 session w2t1p0

Specify one with:
  agent-pro grok session focus <id> --index <0|1>
```

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	if resp.ListITermCalls != 1 || len(resp.Focused) != 0 {
		t.Fatalf("listITerm=%d focused=%v", resp.ListITermCalls, resp.Focused)
	}
	assertErrorOutput(t, resp, `---
version: 3
---
multiple iTerm2 tabs host session `+req.SessionID+`

  \[0\] window 1 \("credit-pricing"\) tab 2 tty /dev/ttys148 session w0t2p0
  \[1\] window 3 \("worktrees"\) tab 1 tty /dev/ttys149 session w2t1p0

Specify one with:
  agent-pro grok session focus `+req.SessionID+` --index <0\|1>
`)
}
```

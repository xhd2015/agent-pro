## Expected

- An out-of-range index is fatal, does not focus, and includes both valid candidates.

## Expected Output

```text
--index 2 is out of range (valid indexes: 0|1)

  [0] window 1 ("credit-pricing") tab 2 tty /dev/ttys148 session w0t2p0
  [1] window 3 ("worktrees") tab 1 tty /dev/ttys149 session w2t1p0
```

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	if resp.ListITermCalls != 1 || len(resp.Focused) != 0 {
		t.Fatalf("listITerm=%d focused=%v", resp.ListITermCalls, resp.Focused)
	}
	assertErrorOutput(t, resp, `---
version: 3
---
--index 2 is out of range \(valid indexes: 0\|1\)

  \[0\] window 1 \("credit-pricing"\) tab 2 tty /dev/ttys148 session w0t2p0
  \[1\] window 3 \("worktrees"\) tab 1 tty /dev/ttys149 session w2t1p0
`)
}
```

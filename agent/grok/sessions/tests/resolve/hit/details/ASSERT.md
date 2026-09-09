## Expected

- Bare id then title/cwd/kind/last active on stdout; stderr empty.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	assertStdoutExact(t, resp.Stdout,
		fixtureSessionID,
		"title:        "+fixtureDetailsTitle,
		"cwd:          "+fixtureDetailsCWD,
		"kind:         main",
		"last active:  2m ago",
	)
	if resp.Stderr != "" {
		t.Fatalf("stderr want empty, got %q", resp.Stderr)
	}
}
```

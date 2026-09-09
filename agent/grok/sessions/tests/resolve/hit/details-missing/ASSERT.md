## Expected

- Bare id only on stdout when summary is missing; exit 0.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	assertStdoutExact(t, resp.Stdout, fixtureSessionID)
	if resp.Stderr != "" {
		t.Fatalf("stderr want empty, got %q", resp.Stderr)
	}
}
```

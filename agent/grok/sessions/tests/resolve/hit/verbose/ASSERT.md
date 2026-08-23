## Expected

- Bare id on stdout; verbose fields on stderr.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	assertStdoutExact(t, resp.Stdout, fixtureSessionID)
	assertStdoutExact(t, resp.Stderr,
		"start pid:    6000",
		"ancestor pid: 4242",
		"runner pid:   4242",
		"source:       open-files",
		"confidence:   hard",
	)
}
```

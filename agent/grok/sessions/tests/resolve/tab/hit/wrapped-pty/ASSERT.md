## Expected

- Bare stdout is the wrapped grok session id (matched via shell ancestor TTY).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	assertStdoutExact(t, resp.Stdout, fixtureTabSessionID)
}
```

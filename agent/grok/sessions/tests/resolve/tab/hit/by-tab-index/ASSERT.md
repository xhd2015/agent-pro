## Expected

- Bare stdout matches tab-2 / index-1 session id.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertOK(t, resp)
	assertStdoutExact(t, resp.Stdout, fixtureTabSessionID)
}
```

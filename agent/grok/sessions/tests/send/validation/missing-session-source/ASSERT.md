## Expected

- No session source → usage error mentioning --session-id / tab.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "--session-id")
	assertNoSend(t, resp)
}
```

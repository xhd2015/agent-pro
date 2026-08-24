## Expected

- Missing text → usage error; no SendText.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "missing text")
	assertNoSend(t, resp)
}
```

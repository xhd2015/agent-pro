## Expected

- Error contains `pid not found`.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrContains(t, resp, "pid not found")
}
```

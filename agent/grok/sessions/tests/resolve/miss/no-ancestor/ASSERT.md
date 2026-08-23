## Expected

- Error contains `no ancestor grok`.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrContains(t, resp, "no ancestor grok")
}
```

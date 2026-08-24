## Expected

- Unknown session → not found; no SendText.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "not found")
	assertNoSend(t, resp)
}
```

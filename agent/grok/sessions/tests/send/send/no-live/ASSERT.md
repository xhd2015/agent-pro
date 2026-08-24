## Expected

- No host without --open → hard error; no SendText.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "no hosting iTerm tab")
	assertNoSend(t, resp)
	assertNoOpen(t, resp)
}
```

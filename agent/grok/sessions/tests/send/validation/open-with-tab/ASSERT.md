## Expected

- `--open` with `--tab` is rejected.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrorContains(t, resp, "--open")
	assertNoSend(t, resp)
	assertNoOpen(t, resp)
}
```

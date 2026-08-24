## Expected

- Exit success; footer `0 sessions`.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertContains(t, resp.Stdout, "0 sessions", "footer")
}
```

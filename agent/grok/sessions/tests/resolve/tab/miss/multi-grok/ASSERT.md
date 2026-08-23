## Expected

- Error refuses to guess when multiple grok sessions share the tab.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrContains(t, resp, "multiple grok sessions on tab 2")
}
```

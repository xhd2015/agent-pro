## Expected

- Mutual exclusion error.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertErrContains(t, resp, "--pid and --tab/--tab-index cannot be specified together")
}
```

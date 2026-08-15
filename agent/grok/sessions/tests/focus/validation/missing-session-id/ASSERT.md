## Expected

- Missing session id is fatal and discovery is not called.

## Expected Output

```text
expected exactly one session id, got 0 arguments
```

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoITerm(t, resp)
	assertErrorOutput(t, resp, `---
version: 3
---
expected exactly one session id, got 0 arguments
`)
}
```

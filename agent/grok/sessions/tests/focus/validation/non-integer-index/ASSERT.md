## Expected

- A non-integer index is fatal and discovery is not called.

## Expected Output

```text
--index must be an integer
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
--index must be an integer
`)
}
```

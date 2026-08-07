## Expected

- `Text` is exactly `""`.
- `Shown`, `SourceCount`, `OldestDropped`, `BodiesTruncated` are 0.
- `LastMessageID` is `""`.
- `Format` equals `FormatDetailed.Text`.

## Errors

- None from `Run`.

```go
import "testing"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoError(t, err)
	assertResp(t, resp)
	assertEmptyResult(t, resp)
}
```

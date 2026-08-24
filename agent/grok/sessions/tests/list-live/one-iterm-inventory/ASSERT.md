## Expected

- Both hosted rows with ITERM from unified inventory.
- `CaptureInventory` runs exactly once (not once per sid / not ListITerm+Capture).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	_ = req
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertContains(t, resp.Stdout, "2 sessions", "footer")
	assertContains(t, resp.Stdout, fixtureListLiveSID, "sid1")
	assertContains(t, resp.Stdout, fixtureListLiveSID2, "sid2")
	assertContains(t, resp.Stdout, "w=3 t=1", "iterm1")
	assertContains(t, resp.Stdout, "TITLE", "header")
	if resp.CaptureInventoryCalls != 1 {
		t.Fatalf("CaptureInventoryCalls = %d, want 1", resp.CaptureInventoryCalls)
	}
	if resp.ListITermCalls != 0 {
		t.Fatalf("ListITermCalls = %d, want 0 (unified inventory)", resp.ListITermCalls)
	}
}
```

## Expected

- Both session rows present.
- `ListProcs` once, `Lsof` once per grok PID, `ListITerm` once.

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertNoHarnessErr(t, err)
	assertNoError(t, resp)
	assertContains(t, resp.Stdout, "2 sessions", "footer")
	assertContains(t, resp.Stdout, fixtureListLiveSID, "sid1")
	assertContains(t, resp.Stdout, fixtureListLiveSID2, "sid2")

	if resp.ListProcsCalls != 1 {
		t.Fatalf("ListProcsCalls = %d, want 1", resp.ListProcsCalls)
	}
	// SETUP adds only grok runners; one lsof per PID after sharing.
	wantLsof := len(req.Procs)
	if wantLsof == 0 {
		t.Fatal("fixture must include grok procs")
	}
	if resp.LsofCalls != wantLsof {
		t.Fatalf("LsofCalls = %d, want %d (one per grok PID)", resp.LsofCalls, wantLsof)
	}
	if resp.ListITermCalls != 1 {
		t.Fatalf("ListITermCalls = %d, want 1", resp.ListITermCalls)
	}
}
```

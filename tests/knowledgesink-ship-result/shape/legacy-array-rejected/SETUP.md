# Scenario

**Feature**: legacy flat `git_commit_files` string array is rejected

```go
import "testing"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.SeedFiles = map[string]string{"INDEX.md": "#\n"}
	req.ResultJSON = []byte(`{
  "has_new_knowledges": true,
  "git_commit_msg": "docs(kb): legacy",
  "git_branch_name": "tester/2026-03-24-legacy",
  "git_commit_files": ["INDEX.md"]
}`)
	req.ExpectOK = false
	req.ExpectErrSubstr = "must be object"
	return nil
}
```

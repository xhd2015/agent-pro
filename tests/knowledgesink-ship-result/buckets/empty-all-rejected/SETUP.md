# Scenario

**Feature**: all-empty git_commit_files object is rejected

```go
import (
	"encoding/json"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/knowledgesink"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	msg, branch := baseMsgBranch()
	body, err := json.Marshal(knowledgesink.ShipResult{
		HasNewKnowledges: boolPtr(true),
		GitCommitMsg:   msg,
		GitBranchName:  branch,
		GitCommitFiles: knowledgesink.ShipCommitFiles{},
	})
	if err != nil {
		return err
	}
	req.ResultJSON = body
	req.ExpectOK = false
	req.ExpectErrSubstr = "at least one path"
	return nil
}
```

# Scenario

**Feature**: add bucket accepts a new file that exists on disk

```go
import (
	"encoding/json"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/knowledgesink"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	msg, branch := baseMsgBranch()
	req.SeedFiles = map[string]string{"topics/new.md": "# new\n"}
	body, err := json.Marshal(knowledgesink.ShipResult{
		HasNewKnowledges: boolPtr(true),
		GitCommitMsg:  msg,
		GitBranchName: branch,
		GitCommitFiles: knowledgesink.ShipCommitFiles{
			Add: []string{"topics/new.md"},
		},
	})
	if err != nil {
		return err
	}
	req.ResultJSON = body
	req.ExpectOK = true
	return nil
}
```

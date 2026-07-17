package agentrunapi_test

import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
	"github.com/xhd2015/agent-pro/pkgs/agentrunapi"
)

func TestBuildFollowUp_embeddingDriver(t *testing.T) {
	got, err := agentrunapi.BuildFollowUpCommand(agentrunapi.FollowUpOpts{
		Driver: agentdriver.Driver{
			Binary: "/abs/spl",
			Args:   []string{"agent-run"},
		},
		SessionID: "s1",
		Prompt:    "hi",
		Open:      true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "/abs/spl") || !strings.Contains(got, "agent-run") || !strings.Contains(got, " run ") {
		t.Fatalf("got %q", got)
	}
	if !strings.HasPrefix(strings.TrimSpace(got), "/abs/spl") {
		// may be quoted
	}
	if !strings.Contains(got, "--open") {
		t.Fatalf("missing --open: %q", got)
	}
}

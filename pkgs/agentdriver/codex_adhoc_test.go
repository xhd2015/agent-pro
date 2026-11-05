//go:build !windows

package agentdriver_test

import (
	"context"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentdriver"
)

func TestActualReceiverMustAdvertiseConnectionGuard(t *testing.T) {
	for _, tc := range []struct {
		name, output string
		ok           bool
	}{{"new", agentdriver.CodexAdhocGuardFlag, true}, {"old", "old agent-run help", false}} {
		t.Run(tc.name, func(t *testing.T) {
			driver := agentdriver.Driver{Binary: "/bin/sh", Args: []string{"-c", "printf '%s' '" + tc.output + "'", "fixture"}}
			err := agentdriver.VerifyCodexAdhoc(context.Background(), driver)
			if (err == nil) != tc.ok {
				t.Fatalf("receiver check: %v", err)
			}
		})
	}
}

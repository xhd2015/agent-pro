package run

import (
	"errors"
	"testing"

	lessflags "github.com/xhd2015/less-flags"
)

func TestParseRunFlagsFromEnv_parsesMockFlagsOnly(t *testing.T) {
	t.Setenv(RunFlagsEnvVar, "--log-events /tmp/events.jsonl --mock-events-preset=think-message")
	opts, err := ParseRunFlagsFromEnv()
	if err != nil {
		t.Fatalf("ParseRunFlagsFromEnv: %v", err)
	}
	if opts.LogEventsPath != "/tmp/events.jsonl" {
		t.Fatalf("LogEventsPath = %q", opts.LogEventsPath)
	}
	if opts.MockEventsPreset != "think-message" {
		t.Fatalf("MockEventsPreset = %q", opts.MockEventsPreset)
	}
}

func TestParseRunFlagsFromEnv_emptyWhenUnset(t *testing.T) {
	t.Setenv(RunFlagsEnvVar, "")
	opts, err := ParseRunFlagsFromEnv()
	if err != nil {
		t.Fatalf("ParseRunFlagsFromEnv: %v", err)
	}
	if opts != (RunGrokOptions{}) {
		t.Fatalf("opts = %+v, want zero", opts)
	}
}

func TestParseRunFlags_agentRunnerConfigHome(t *testing.T) {
	opts, remain, err := ParseRunFlags([]string{"--agent-runner-config-home", "/tmp/grok-home", "grok"})
	if err != nil {
		t.Fatalf("ParseRunFlags: %v", err)
	}
	if opts.AgentRunnerConfigHome != "/tmp/grok-home" {
		t.Fatalf("AgentRunnerConfigHome = %q", opts.AgentRunnerConfigHome)
	}
	if len(remain) != 1 || remain[0] != "grok" {
		t.Fatalf("remain = %#v", remain)
	}
}

func TestParseRunFlagsFromEnv_help(t *testing.T) {
	t.Setenv(RunFlagsEnvVar, "--help")
	_, err := ParseRunFlagsFromEnv()
	if !errors.Is(err, lessflags.ErrHelp) {
		t.Fatalf("ParseRunFlagsFromEnv err = %v, want ErrHelp", err)
	}
}

func TestParseRunFlags_passesGrokFlagsAfterAgentName(t *testing.T) {
	opts, remain, err := ParseRunFlags([]string{"--log-events", "/tmp/e.jsonl", "grok", "--always-approve", "--model", "m1"})
	if err != nil {
		t.Fatalf("ParseRunFlags: %v", err)
	}
	if opts.LogEventsPath != "/tmp/e.jsonl" {
		t.Fatalf("LogEventsPath = %q", opts.LogEventsPath)
	}
	if len(remain) != 4 || remain[0] != "grok" || remain[1] != "--always-approve" || remain[2] != "--model" {
		t.Fatalf("remain = %#v", remain)
	}
}
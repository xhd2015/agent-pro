package agenttty

import (
	"testing"
)

func TestOpenCloseExits_defaultOn(t *testing.T) {
	t.Setenv("AGENT_RUN_OPEN_CLOSE_EXITS", "")
	if !OpenCloseExits() {
		t.Fatal("empty env: want OpenCloseExits true (production default)")
	}
}

func TestOpenCloseExits_optOut(t *testing.T) {
	for _, v := range []string{"0", "false", "no", "off", "FALSE", " Off "} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("AGENT_RUN_OPEN_CLOSE_EXITS", v)
			if OpenCloseExits() {
				t.Fatalf("env %q: want OpenCloseExits false", v)
			}
		})
	}
}

func TestOpenCloseExits_explicitOn(t *testing.T) {
	for _, v := range []string{"1", "true", "yes", "anything"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("AGENT_RUN_OPEN_CLOSE_EXITS", v)
			if !OpenCloseExits() {
				t.Fatalf("env %q: want OpenCloseExits true", v)
			}
		})
	}
}

func TestOpenCloseExitsExperiment_alias(t *testing.T) {
	t.Setenv("AGENT_RUN_OPEN_CLOSE_EXITS", "0")
	if OpenCloseExitsExperiment() {
		t.Fatal("alias should match OpenCloseExits")
	}
	t.Setenv("AGENT_RUN_OPEN_CLOSE_EXITS", "1")
	if !OpenCloseExitsExperiment() {
		t.Fatal("alias should match OpenCloseExits")
	}
}

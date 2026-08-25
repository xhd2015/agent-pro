package agentruncli

import (
	"strings"
	"testing"

	"github.com/xhd2015/agent-pro/pkgs/agentstorage"
)

func TestResumeHelpMentionsModelReasoningEffort(t *testing.T) {
	if !strings.Contains(resumeHelp, "--model-reasoning-effort") {
		t.Fatal("resume help must document --model-reasoning-effort pass-through")
	}
}

func TestResolveResumeModelAndEffort(t *testing.T) {
	meta := agentstorage.SessionMeta{Model: "gpt-5.6-luna"}

	model, effort := resolveResumeModelAndEffort(resumeRunConfig{
		model:                "gpt-5.6-terra",
		modelReasoningEffort: "max",
	}, meta)
	if model != "gpt-5.6-terra" || effort != "max" {
		t.Fatalf("cli wins: model=%q effort=%q", model, effort)
	}

	model, effort = resolveResumeModelAndEffort(resumeRunConfig{
		modelReasoningEffort: "high",
	}, meta)
	if model != "gpt-5.6-luna" || effort != "high" {
		t.Fatalf("meta model + cli effort: model=%q effort=%q", model, effort)
	}

	model, effort = resolveResumeModelAndEffort(resumeRunConfig{}, meta)
	if model != "gpt-5.6-luna" || effort != "" {
		t.Fatalf("empty effort must stay empty (no invent): model=%q effort=%q", model, effort)
	}
}

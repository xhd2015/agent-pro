package agentui

import (
	"fmt"
	"strings"

	lessflags "github.com/xhd2015/less-flags"
)

type runOptions struct {
	Model       string
	OutputFile  string
	ResumeID    string
	AgentRunner string
	Feature     string
}

func parseRunOptions(cfg Config, args []string) (runOptions, error) {
	var modelFlag *string
	var outputFlag *string
	var resumeFlag *string
	var agentRunnerFlag *string

	remainArgs, err := lessflags.String("--model", &modelFlag).
		String("-o,--output", &outputFlag).
		String("--resume", &resumeFlag).
		String("--agent-runner", &agentRunnerFlag).
		Help("-h,--help", cfg.Usage).
		Parse(args)
	if err != nil {
		return runOptions{}, err
	}

	agentRunner := "opencode"
	if agentRunnerFlag != nil {
		agentRunner = *agentRunnerFlag
	}
	if agentRunner != "opencode" && agentRunner != "codex" {
		return runOptions{}, fmt.Errorf("unsupported agent runner: %s (supported: opencode, codex)", agentRunner)
	}
	if agentRunner == "codex" {
		return runOptions{}, fmt.Errorf("codex runner not yet implemented")
	}

	llmModel := ""
	if modelFlag != nil {
		llmModel = *modelFlag
	}
	outputFile := ""
	if outputFlag != nil {
		outputFile = *outputFlag
	}
	resumeID := ""
	if resumeFlag != nil {
		resumeID = *resumeFlag
	}
	feature := strings.Join(remainArgs, " ")

	return runOptions{
		Model:       llmModel,
		OutputFile:  outputFile,
		ResumeID:    resumeID,
		AgentRunner: agentRunner,
		Feature:     feature,
	}, nil
}

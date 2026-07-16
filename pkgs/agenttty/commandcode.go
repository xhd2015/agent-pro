package agenttty

import (
	"fmt"
	"os"
	"strings"

	agentexec "github.com/xhd2015/agent-pro/agent/exec"
)

const envCommandcodeTTYCommand = "AGENT_RUN_COMMANDCODE_TTY_COMMAND"

func BuildCommandcodeCommandArgv(env *agentexec.Env, settingsPath, agentRunnerBinary, model, resumeSession string) ([]string, error) {
	if hook := strings.TrimSpace(os.Getenv(envCommandcodeTTYCommand)); hook != "" {
		return parseShellWords(hook)
	}
	spec := strings.TrimSpace(agentRunnerBinary)
	if spec == "" {
		return nil, fmt.Errorf("commandcode-tty: --agent-runner-binary is required (e.g. llm-mock-run-commandcode)")
	}
	words, err := parseShellWords(spec)
	if err != nil {
		return nil, err
	}
	return words, nil
}

func detectCommandcodeScreenStatus(scrollback []byte) string {
	plain := stripPlain(scrollback)
	if strings.TrimSpace(plain) == "" {
		return "starting"
	}
	return "idle"
}

func checkCommandcodeWritable(scrollback []byte) WritableStatus {
	return WritableStatus{Ready: true}
}

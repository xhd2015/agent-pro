package subagent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/logs"

	"github.com/xhd2015/agent-pro/agent/event/logging"
)

type Config struct {
	RoleName         string
	Cmd              string
	PromptContent    string
	SessionEnvVar    string
	SessionMetaField string
	DebugSessionEnv  string
	AgentRunnerEnv   string
	ModelEnv         string
}

type Options struct {
	Prompt       string
	AgentRunner  string
	MockConfig   string
	SessionID    string
	Requirement  string
	CatchUp      bool
	Status       bool
	ListSessions bool
	SessionBase  string
}

func (c Config) agentPrompt() string {
	s := c.PromptContent
	if strings.HasPrefix(s, "---\n") {
		rest := s[4:]
		if idx := strings.Index(rest, "\n---\n"); idx >= 0 {
			s = rest[idx+5:]
			if strings.HasPrefix(s, "\n") {
				s = s[1:]
			}
		}
	}
	return s
}

func (c Config) sessionEnvVar() string {
	if c.SessionEnvVar != "" {
		return c.SessionEnvVar
	}
	return "AGENT_PRO_SUBAGENT_" + strings.ToUpper(c.RoleName) + "_SESSION_ID"
}

func (c Config) metaSessionField() string {
	if c.SessionMetaField != "" {
		return c.SessionMetaField
	}
	return "subagent_role_" + c.RoleName + "_session_id"
}

// agentRunnerEnv returns the env var name used to override the auto-detected
// agent runner. Returns empty string when not configured, which disables
// the env-override feature.
// Consumers (e.g. doctest) should configure this via Config.AgentRunnerEnv.
func (c Config) agentRunnerEnv() string {
	return c.AgentRunnerEnv
}

// modelEnv returns the env var name used to override the model passed to the
// agent runner. Returns empty string when not configured.
// Consumers (e.g. doctest) should configure this via Config.ModelEnv.
func (c Config) modelEnv() string {
	return c.ModelEnv
}

func PromptContent(c Config) string {
	return c.PromptContent
}

func Run(ctx context.Context, c Config, opts Options) error {
	if opts.ListSessions {
		if opts.SessionID != "" {
			fmt.Fprintf(os.Stderr, "error: --list-sessions and --session-id are mutually exclusive\n")
			return nil
		}
		return listSessions(c, opts)
	}

	if opts.Status {
		if opts.SessionID == "" && os.Getenv(c.sessionEnvVar()) == "" && os.Getenv("CODEX_THREAD_ID") == "" {
			fmt.Fprintf(os.Stderr, "error: --status requires --session-id\n")
			return nil
		}
		return showStatus(c, opts)
	}

	if opts.CatchUp {
		if opts.SessionID == "" && os.Getenv(c.sessionEnvVar()) == "" && os.Getenv("CODEX_THREAD_ID") == "" {
			return fmt.Errorf("--trace requires --session-id")
		}
		return traceSession(c, opts)
	}

	prompt := strings.TrimSpace(opts.Prompt)
	if opts.Requirement != "" {
		data, err := os.ReadFile(opts.Requirement)
		if err != nil {
			return fmt.Errorf("read requirement file %s: %w", opts.Requirement, err)
		}
		reqContent := strings.TrimSpace(string(data))
		if prompt != "" {
			prompt = reqContent + "\n\n---\n\n" + prompt
		} else {
			prompt = reqContent
		}
	}
	if prompt == "" {
		return fmt.Errorf("agent %s requires <prompt>", c.RoleName)
	}

	// If no explicit agent runner is provided, try to auto-detect.
	// If we can't detect an agent context, fail with a clear error.
	agentRunner := strings.TrimSpace(opts.AgentRunner)
	if agentRunner == "" {
		runner, ok := autoDetectAgentRunner(c)
		if !ok {
			return fmt.Errorf("agent %s cannot resolve session: must be run inside an agent session (opencode, pi, codex, or crush)", c.RoleName)
		}
		agentRunner = runner
	}

	// Validate the agent runner before any session setup so callers get
	// a clear "unknown agent runner" error rather than a confusing
	// "cannot detect session" error.
	validRunners := map[string]bool{
		"opencode": true, "codex": true, "pi": true, "crush": true,
		"cursor": true, "fake-codex": true,
	}
	if !validRunners[agentRunner] {
		return fmt.Errorf("unknown agent runner id: %s", agentRunner)
	}

	model := os.Getenv(c.modelEnv())

	srcs, err := resolveSessionID(c, opts.SessionID)
	if err != nil {
		return err
	}
	srcs.agentRunner = agentRunner

	sessionDir, isNew, err := findOrCreateSession(c, opts, srcs.sessionID, srcs)
	if isNew {
		Logf("Session created: %s (source: %s)\n", srcs.sessionID, sourceLabel(srcs))
	} else {
		Logf("Session resumed: %s (source: %s)\n", srcs.sessionID, sourceLabel(srcs))
	}
	if err != nil {
		return fmt.Errorf("session: %w", err)
	}

	if err := writeSessionPID(sessionDir); err != nil {
		return fmt.Errorf("write pid: %w", err)
	}
	defer removeSessionPID(sessionDir)

	msgPath := filepath.Join(sessionDir, "messages.jsonl")
	msgEntry := map[string]string{
		"type":        "message",
		"content":     prompt,
		"create_time": time.Now().Format(time.RFC3339),
	}
	msgData, _ := json.Marshal(msgEntry)
	f, err := os.OpenFile(msgPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("write messages.jsonl: %w", err)
	}
	fmt.Fprintf(f, "%s\n", string(msgData))
	f.Close()

	tempDir, err := os.MkdirTemp("", "agent-pro-agent-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tempDir)

	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable: %w", err)
	}
	ypqPath := filepath.Join(tempDir, "yield-pending-questions")
	if out, err := exec.Command("cp", exe, ypqPath).CombinedOutput(); err != nil {
		return fmt.Errorf("copy yield-pending-questions: %w\n%s", err, string(out))
	}

	pathEntry := tempDir + string(filepath.ListSeparator)
	os.Setenv("PATH", pathEntry+os.Getenv("PATH"))

	questionsDir := filepath.Join(sessionDir, "questions")
	if err := os.MkdirAll(questionsDir, 0755); err != nil {
		return fmt.Errorf("create questions dir: %w", err)
	}
	questionFile := newQuestionsFile(questionsDir)
	os.Setenv("QUESTION_FIFO", questionFile)

	progressDir := filepath.Join(sessionDir, "progress")
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		return fmt.Errorf("create progress dir: %w", err)
	}

	rpPath := filepath.Join(tempDir, "report-progress")
	if out, err := exec.Command("cp", exe, rpPath).CombinedOutput(); err != nil {
		return fmt.Errorf("copy report-progress: %w\n%s", err, string(out))
	}

	progressFile := filepath.Join(progressDir, time.Now().Format("20060102_150405")+"_progress_update.jsonl")
	os.Setenv("PROGRESS_FILE", progressFile)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		logs.WatchLine(ctx, progressFile, logs.WatchLineOptions{}, func(line string) error {
			var entry map[string]any
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				return nil
			}
			desc, _ := entry["description"].(string)
			Logf("%s", desc)
			return nil
		})
	}()

	if opts.MockConfig != "" {
		os.Setenv("FAKE_CODEX_MOCK_CONFIG", opts.MockConfig)
	}

	var opencodeSessionID string
	if !isNew {
		opencodeSessionID = readOpencodeSessionID(sessionDir)
	}

	eventsPath := filepath.Join(sessionDir, "events.jsonl")
	eventsLogger, err := logging.Open(eventsPath)
	if err != nil {
		return fmt.Errorf("open events.jsonl: %w", err)
	}
	defer eventsLogger.Close()

	capture := &sessionLogWriter{
		eventsFile:  eventsLogger,
		agentRunner: agentRunner,
	}

	var fullPrompt string
	if isNew {
		fullPrompt = c.agentPrompt() + "\n\n---\n\n<user_request>\n" + prompt + "\n</user_request>\n"
	} else {
		fullPrompt = prompt
	}
	output, err := runAgent(ctx, agentRunner, model, fullPrompt, opencodeSessionID, capture)
	cancel()
	if err != nil {
		return fmt.Errorf("sub-agent failed: %w", err)
	}

	if isNew {
		sid := capture.sessionID
		if sid == "" {
			sid = srcs.sessionID
		}
		if updateErr := updateSessionMeta(sessionDir, sid, srcs); updateErr != nil {
			return fmt.Errorf("update session meta: %w", updateErr)
		}
	}

	fmt.Print(output)

	f, fErr := os.Open(questionFile)
	if fErr == nil {
		defer f.Close()
		var buf bytes.Buffer
		buf.ReadFrom(f)
		if buf.Len() > 0 {
			fmt.Print("\n\n---\nQUESTIONS\n---\n\n")
			fmt.Print(buf.String())
		}
	}

	return nil
}

func autoDetectAgentRunner(c Config) (runner string, detected bool) {
	// Priority 1: Env var override
	if v := os.Getenv(c.agentRunnerEnv()); v != "" {
		return strings.TrimSpace(v), true
	}
	// Priority 2: CODEX_THREAD_ID detection
	if v := os.Getenv("CODEX_THREAD_ID"); v != "" {
		return "codex", true
	}
	// Priority 3: PI_CODING_AGENT env var (set by pi TUI)
	if v := os.Getenv("PI_CODING_AGENT"); v != "" {
		return "pi", true
	}
	// Priority 4: Parent process detection
	if ppid := os.Getppid(); ppid > 0 {
		comm := getProcessName(ppid)
		if comm != "" {
			switch strings.ToLower(comm) {
			case "opencode":
				return "opencode", true
			case "pi":
				return "pi", true
			case "crush":
				return "crush", true
			case "codex":
				return "codex", true
			}
		}
		// Grandparent walk for pi only (pi spawns a shell which spawns doctest)
		if pppid := getParentPid(ppid); pppid > 0 {
			if pcomm := getProcessName(pppid); strings.ToLower(pcomm) == "pi" {
				return "pi", true
			}
		}
	}
	return "", false
}

func getProcessName(pid int) string {
	if TestProcessNameFunc != nil {
		return TestProcessNameFunc(pid)
	}
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("ps", "-o", "comm=", "-p", fmt.Sprintf("%d", pid)).Output()
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(out))
	}
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid))
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(data))
	}
	return ""
}

// getParentPid returns the parent PID of the given PID.
func getParentPid(pid int) int {
	if runtime.GOOS == "darwin" {
		out, err := exec.Command("ps", "-o", "ppid=", "-p", fmt.Sprintf("%d", pid)).Output()
		if err != nil {
			return -1
		}
		ppidStr := strings.TrimSpace(string(out))
		var ppid int
		if _, err := fmt.Sscanf(ppidStr, "%d", &ppid); err != nil {
			return -1
		}
		return ppid
	}
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			return -1
		}
		// /proc/PID/stat format: PID (comm) state ppid ...
		// We need the 4th field (ppid). The comm field may contain spaces inside ().
		s := string(data)
		// Find the closing ')' of the comm field, the ppid is right after.
		closeParen := strings.LastIndex(s, ")")
		if closeParen < 0 {
			return -1
		}
		rest := strings.TrimSpace(s[closeParen+1:])
		// rest starts with " state ppid ..."
		fields := strings.Fields(rest)
		if len(fields) < 2 {
			return -1
		}
		var ppid int
		if _, err := fmt.Sscanf(fields[1], "%d", &ppid); err != nil {
			return -1
		}
		return ppid
	}
	return -1
}

func Logf(fmtStr string, args ...interface{}) {
	ts := time.Now().Format("2006-01-02T15:04:05")
	s := fmt.Sprintf(fmtStr, args...)
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	fmt.Print("[" + ts + "] " + s)
}

package subagent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
	// SessionRetryHint formats the suggested retry CLI when session id cannot be
	// auto-detected. Receives sessionID and prompt. When nil, defaults to
	// `doctest agent <Cmd> --session-id <id> <prompt>`.
	SessionRetryHint func(sessionID, prompt string) string
	// AutoGenerateSessionID when true generates a session ID when flag, env var,
	// and CODEX_THREAD_ID are all unset. Default false returns an error with retry hint.
	AutoGenerateSessionID bool
}

type AgentRunInfo struct {
	InnerSessionID string
	AgentRunner    string
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
	SessionLayout SessionLayout
	Timeout      time.Duration
	// StdoutWriter receives streamed agent output instead of os.Stdout when set.
	StdoutWriter io.Writer

	// When true, subagent must not read or write meta.json. Session match/resume
	// uses SessionID and ResumeInnerSessionID; host persists via OnAgentComplete.
	HostOwnsMeta bool
	// Inner runner session for resume (from host meta.opencode_session_id).
	ResumeInnerSessionID string
	// Called after successful agent run; host persists to its meta.
	OnAgentComplete func(AgentRunInfo) error
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
		if opts.SessionID == "" && os.Getenv(c.sessionEnvVar()) == "" && os.Getenv("CODEX_THREAD_ID") == "" && !c.AutoGenerateSessionID {
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
		runner, ok := AutoDetectAgentRunner(c)
		if !ok {
			return fmt.Errorf("agent %s cannot resolve session: must be run inside an agent session (opencode, pi, codex, crush, or grok)", c.RoleName)
		}
		agentRunner = runner
	}

	// Validate the agent runner before any session setup so callers get
	// a clear "unknown agent runner" error rather than a confusing
	// "cannot detect session" error.
	validRunners := map[string]bool{
		"opencode": true, "codex": true, "pi": true, "crush": true,
		"grok": true, "cursor": true, "fake-codex": true,
	}
	if !validRunners[agentRunner] {
		return fmt.Errorf("unknown agent runner id: %s", agentRunner)
	}

	model := os.Getenv(c.modelEnv())

	srcs, err := resolveSessionID(c, opts.SessionID, prompt)
	if err != nil {
		return err
	}
	srcs.agentRunner = agentRunner

	layout := opts.SessionLayout
	sessionDir, isNew, err := findOrCreateSession(c, opts, srcs.sessionID, srcs)
	if isNew {
		Logf("Session created: %s (source: %s)\n", srcs.sessionID, sourceLabel(srcs))
	} else {
		Logf("Session resumed: %s (source: %s)\n", srcs.sessionID, sourceLabel(srcs))
	}
	if err != nil {
		return fmt.Errorf("session: %w", err)
	}

	paths := resolvedSessionPaths(sessionDir, layout)
	if err := writeSessionPIDAt(paths.pidPath); err != nil {
		return fmt.Errorf("write pid: %w", err)
	}
	defer removeSessionPIDAt(paths.pidPath)

	if !layout.skipMessages() {
		msgPath := paths.messagesPath
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
	}

	var questionFile string
	var progressFile string
	if layout.questionsEnabled() || layout.progressEnabled() {
		tempDir, err := os.MkdirTemp("", "agent-pro-agent-*")
		if err != nil {
			return fmt.Errorf("create temp dir: %w", err)
		}
		defer os.RemoveAll(tempDir)

		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("get executable: %w", err)
		}
		pathEntry := tempDir + string(filepath.ListSeparator)
		os.Setenv("PATH", pathEntry+os.Getenv("PATH"))

		if layout.questionsEnabled() {
			ypqPath := filepath.Join(tempDir, "yield-pending-questions")
			if out, err := exec.Command("cp", exe, ypqPath).CombinedOutput(); err != nil {
				return fmt.Errorf("copy yield-pending-questions: %w\n%s", err, string(out))
			}
			questionsDir := paths.questionsDir
			if err := os.MkdirAll(questionsDir, 0755); err != nil {
				return fmt.Errorf("create questions dir: %w", err)
			}
			questionFile = newQuestionsFile(questionsDir)
			os.Setenv("QUESTION_FIFO", questionFile)
		} else {
			os.Unsetenv("QUESTION_FIFO")
		}

		if layout.progressEnabled() {
			rpPath := filepath.Join(tempDir, "report-progress")
			if out, err := exec.Command("cp", exe, rpPath).CombinedOutput(); err != nil {
				return fmt.Errorf("copy report-progress: %w\n%s", err, string(out))
			}
			progressDir := paths.progressDir
			if err := os.MkdirAll(progressDir, 0755); err != nil {
				return fmt.Errorf("create progress dir: %w", err)
			}
			progressFile = filepath.Join(progressDir, time.Now().Format("20060102_150405")+"_progress_update.jsonl")
			os.Setenv("PROGRESS_FILE", progressFile)
		} else {
			os.Unsetenv("PROGRESS_FILE")
		}
	} else {
		os.Unsetenv("QUESTION_FIFO")
		os.Unsetenv("PROGRESS_FILE")
	}

	if opts.Timeout > 0 {
		var timeoutCancel context.CancelFunc
		ctx, timeoutCancel = context.WithTimeout(ctx, opts.Timeout)
		defer timeoutCancel()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if layout.progressEnabled() && progressFile != "" {
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
	}

	if opts.MockConfig != "" {
		os.Setenv("FAKE_CODEX_MOCK_CONFIG", opts.MockConfig)
	}

	var opencodeSessionID string
	if !isNew {
		if opts.HostOwnsMeta {
			opencodeSessionID = opts.ResumeInnerSessionID
		} else {
			opencodeSessionID = readOpencodeSessionIDAt(paths.metaPath)
		}
	}

	eventsPath := paths.eventsPath
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
	output, err := runAgent(ctx, agentRunner, model, fullPrompt, opencodeSessionID, capture, opts.StdoutWriter)
	cancel()
	if err != nil {
		return fmt.Errorf("sub-agent failed: %w", err)
	}

	innerID := resolveInnerSessionID(capture.sessionID, isNew, srcs)
	if innerID != "" {
		if opts.HostOwnsMeta {
			if opts.OnAgentComplete != nil {
				if cbErr := opts.OnAgentComplete(AgentRunInfo{
					InnerSessionID: innerID,
					AgentRunner:    agentRunner,
				}); cbErr != nil {
					return fmt.Errorf("on agent complete: %w", cbErr)
				}
			}
		} else {
			current := readOpencodeSessionIDAt(paths.metaPath)
			if isNew || current == "" {
				if updateErr := updateSessionMetaAt(paths.metaPath, innerID, srcs); updateErr != nil {
					return fmt.Errorf("update session meta: %w", updateErr)
				}
			}
		}
	}

	writeStdout(opts.StdoutWriter, output)

	if layout.questionsEnabled() && questionFile != "" {
		f, fErr := os.Open(questionFile)
		if fErr == nil {
			defer f.Close()
			var buf bytes.Buffer
			buf.ReadFrom(f)
			if buf.Len() > 0 {
				writeStdout(opts.StdoutWriter, "\n\n---\nQUESTIONS\n---\n\n")
				writeStdout(opts.StdoutWriter, buf.String())
			}
		}
	}

	return nil
}

func writeStdout(w io.Writer, s string) {
	if w != nil {
		_, _ = w.Write([]byte(s))
		return
	}
	fmt.Print(s)
}

func Logf(fmtStr string, args ...interface{}) {
	ts := time.Now().Format("2006-01-02T15:04:05")
	s := fmt.Sprintf(fmtStr, args...)
	if !strings.HasSuffix(s, "\n") {
		s += "\n"
	}
	fmt.Print("[" + ts + "] " + s)
}

package agenttty

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strings"
	"time"

	types "github.com/xhd2015/agent-pro/agent/event/types"
	agentexec "github.com/xhd2015/agent-pro/agent/exec"
)

const (
	envStubTTYCommand  = "AGENT_RUN_STUB_TTY_COMMAND"
	envStubTTYScenario = "AGENT_RUN_STUB_TTY_SCENARIO"
)

// StubScenario configures the test-only stub TTY mock process.
type StubScenario struct {
	BannerDelayMs       int         `json:"banner_delay_ms"`
	BannerText          string      `json:"banner_text"`
	PromptLatencyMs     int         `json:"prompt_latency_ms"`
	ResponseText        string      `json:"response_text"`
	ScreenStatus        string      `json:"screen_status"`
	ScreenFrames        []StubFrame `json:"screen_frames"`
	LLMEvents           []StubEvent `json:"llm_events"`
	RunnerSessionID     string      `json:"runner_session_id"`
	TurnCompleteDelayMs int         `json:"turn_complete_delay_ms"`
	ExitAfterTurn       bool        `json:"exit_after_turn"`
	WritableReason      string      `json:"writable_reason"`
}

type StubFrame struct {
	DelayMs int    `json:"delay_ms"`
	Text    string `json:"text"`
}

type StubEvent struct {
	DelayMs int    `json:"delay_ms"`
	Type    string `json:"type"`
	Role    string `json:"role"`
	Text    string `json:"text"`
}

var cachedStubScenario *StubScenario

func loadStubScenario() *StubScenario {
	if cachedStubScenario != nil {
		return cachedStubScenario
	}
	path := strings.TrimSpace(os.Getenv(envStubTTYScenario))
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var sc StubScenario
	if json.Unmarshal(data, &sc) != nil {
		return nil
	}
	cachedStubScenario = &sc
	return cachedStubScenario
}

// BuildStubCommandArgv returns argv for the stub TTY process inside the PTY.
func BuildStubCommandArgv(env *agentexec.Env, settingsPath, agentPath, model, resumeSession string) ([]string, error) {
	if hook := strings.TrimSpace(os.Getenv(envStubTTYCommand)); hook != "" {
		return parseShellWords(hook)
	}
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	return []string{exe, "__stub-tty"}, nil
}

func tailStubLLMEvents(ctx TailContext) (string, error) {
	sc := loadStubScenarioFromPath(ctx.ScenarioPath)
	if sc == nil || ctx.Emit == nil {
		return "", nil
	}
	start := time.Now()
	for _, ev := range sc.LLMEvents {
		delay := time.Duration(ev.DelayMs) * time.Millisecond
		elapsed := time.Since(start)
		if delay > elapsed {
			time.Sleep(delay - elapsed)
		}
		switch ev.Type {
		case "message", "assistant":
			_ = ctx.Emit(types.AgentEvent{
				Type:      types.ActionMessage,
				Role:      firstNonEmpty(ev.Role, "assistant"),
				Text:      ev.Text,
				Timestamp: time.Now().UnixMilli(),
			})
		case "done":
			_ = ctx.Emit(types.AgentEvent{Type: types.ActionDone, Timestamp: time.Now().UnixMilli()})
		default:
			if ev.Text != "" {
				_ = ctx.Emit(types.AgentEvent{
					Type:      types.ActionMessage,
					Role:      firstNonEmpty(ev.Role, "assistant"),
					Text:      ev.Text,
					Timestamp: time.Now().UnixMilli(),
				})
			}
		}
	}
	return sc.RunnerSessionID, nil
}

func loadStubScenarioFromPath(path string) *StubScenario {
	if path == "" {
		return loadStubScenario()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var sc StubScenario
	if json.Unmarshal(data, &sc) != nil {
		return nil
	}
	return &sc
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// RunStubTTYMain runs the test-only stub TTY alt-screen mock (invoked via __stub-tty).
func RunStubTTYMain() error {
	sc := loadStubScenario()
	if sc == nil {
		sc = &StubScenario{
			BannerText:    "STUB_TTY_BANNER",
			ScreenStatus:  "idle",
			ExitAfterTurn: true,
			ScreenFrames: []StubFrame{
				{DelayMs: 0, Text: "STUB_TTY_BANNER\n\u203a "},
			},
		}
	}
	if sc.BannerText == "" {
		sc.BannerText = "STUB_TTY_BANNER"
	}

	start := time.Now()
	currentText := ""
	framesDone := make(chan struct{})

	go func() {
		defer close(framesDone)
		for _, frame := range sc.ScreenFrames {
			wait := time.Duration(frame.DelayMs)*time.Millisecond - time.Since(start)
			if wait > 0 {
				time.Sleep(wait)
			}
			if frame.Text != "" {
				currentText = frame.Text
				writeScreen(frame.Text)
			}
		}
		if len(sc.ScreenFrames) == 0 {
			if sc.BannerDelayMs > 0 {
				time.Sleep(time.Duration(sc.BannerDelayMs) * time.Millisecond)
			}
			text := sc.BannerText + "\n\u203a "
			currentText = text
			writeScreen(text)
		}
	}()

	reader := bufio.NewReader(os.Stdin)
	if !sc.ExitAfterTurn {
		<-framesDone
		go stubEchoLoop(reader, &currentText)
		select {}
	}

	<-framesDone
	// Headless new sessions put the prompt on argv (Grok-style) and do not
	// re-inject. Auto-complete the turn when argv has a prompt after
	// __stub-tty; otherwise wait briefly for an injected newline.
	shouldTurn := stubArgvHasPrompt() || waitForStubInput(reader, 3*time.Second)
	if shouldTurn {
		if sc.PromptLatencyMs > 0 {
			time.Sleep(time.Duration(sc.PromptLatencyMs) * time.Millisecond)
		}
		response := sc.ResponseText
		if response == "" {
			response = "Response: mock assistant reply"
		}
		newText := sc.BannerText + "\n\u203a prompt\n" + response + "\n\u203a "
		currentText = newText
		writeScreen(newText)
		if sc.TurnCompleteDelayMs > 0 {
			time.Sleep(time.Duration(sc.TurnCompleteDelayMs) * time.Millisecond)
		}
	}
	return nil
}

// stubArgvHasPrompt reports whether headless attached a trailing prompt after __stub-tty.
func stubArgvHasPrompt() bool {
	for i, a := range os.Args {
		if a == "__stub-tty" {
			if i+1 < len(os.Args) && strings.TrimSpace(os.Args[i+1]) != "" {
				return true
			}
			return false
		}
	}
	return false
}

func waitForStubInput(reader *bufio.Reader, timeout time.Duration) bool {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	type result struct {
		ok bool
	}
	ch := make(chan result, 1)
	go func() {
		for {
			b, err := reader.ReadByte()
			if err != nil {
				ch <- result{ok: false}
				return
			}
			if b == '\n' || b == '\r' {
				ch <- result{ok: true}
				return
			}
		}
	}()
	select {
	case r := <-ch:
		return r.ok
	case <-time.After(timeout):
		return false
	}
}

func stubEchoLoop(reader *bufio.Reader, currentText *string) {
	var line strings.Builder
	for {
		b, err := reader.ReadByte()
		if err != nil {
			return
		}
		if b == '\n' || b == '\r' {
			trimmed := strings.TrimSpace(line.String())
			line.Reset()
			if trimmed == "" {
				continue
			}
			updated := *currentText + trimmed + "\n"
			*currentText = updated
			writeScreen(updated)
			continue
		}
		line.WriteByte(b)
	}
}

func writeScreen(text string) {
	_, _ = io.WriteString(os.Stdout, "\x1b[2J\x1b[H"+text)
}
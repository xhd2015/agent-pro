package agenttty

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/tty-watch/pkgs/ttywatch"
)

const (
	// codexSubmitConfirmInterval is the pause before the first extra Enter and
	// between later ones, so a slow rollout flush is not a second submit.
	codexSubmitConfirmInterval = time.Second
	// codexSubmitConfirmBudget caps how long a dropped Enter is retried.
	codexSubmitConfirmBudget = 2 * time.Minute
	// codexSubmitOverlapMin is the agreed "mostly the same" bar. Codex may
	// store a slightly edited copy of the typed prompt.
	codexSubmitOverlapMin = 0.80
	// codexSubmitExactMaxBytes is the size above which overlap uses 64-rune
	// chunks instead of a full edit matrix.
	codexSubmitExactMaxBytes = 32 * 1024
	codexSubmitChunkRunes    = 64
	codexSubmitChunkSlack    = 64
	// codexSubmitLineMax is the largest rollout line we parse. The older
	// prompt matcher stopped at 512KB and missed long SeaTalk injects.
	codexSubmitLineMax = 16 << 20
)

// CodexSubmitConfirmOpts drives a rollout-backed Enter retry.
// Baselines, when non-nil, are file sizes captured before the inject Enter.
// A nil map is captured at start (too late if the turn already flushed).
type CodexSubmitConfirmOpts struct {
	ListenAddr        string
	TerminalSessionID string
	Prompt            string
	Runner            string
	CodexHome         string
	Workspace         string
	Baselines         map[string]int64
}

// StartCodexSubmitConfirm blocks until the injected prompt shows up as a user
// message in a cwd-matched Codex rollout, the budget elapses, or ctx is done.
// Until then it sends one bare Enter per interval. Call it in a goroutine.
// NoSubmit callers must not start it.
func StartCodexSubmitConfirm(ctx context.Context, opts CodexSubmitConfirmOpts) {
	if ctx == nil {
		ctx = context.Background()
	}
	opts = opts.normalized()
	if opts.Prompt == "" || strings.TrimSpace(opts.ListenAddr) == "" || strings.TrimSpace(opts.TerminalSessionID) == "" {
		return
	}
	if opts.Baselines == nil {
		opts.Baselines = CodexRolloutBaselines(opts.CodexHome, opts.Workspace)
	}
	listenAddr := opts.ListenAddr
	sessionID := opts.TerminalSessionID
	runCodexSubmitConfirmLoop(ctx, codexSubmitLoop{
		interval: codexSubmitConfirmInterval,
		budget:   codexSubmitConfirmBudget,
		after:    time.After,
		landed: func() bool {
			return codexPromptLanded(opts.CodexHome, opts.Workspace, opts.Prompt, opts.Baselines)
		},
		enter: func() error {
			return ttywatch.SendMessage(listenAddr, sessionID, "", true)
		},
	})
}

// CodexSubmitPlace reads the agent-run session bound to a live terminal so a
// follow-up drain can poll the same Codex home and workspace the TUI used.
// Missing meta falls back to CodexHome() and an empty workspace.
func CodexSubmitPlace(agentRunHome, runner, terminalSessionID string) (codexHome, workspace string) {
	codexHome = CodexHome()
	agentID := findAgentSessionByTerminalID(agentRunHome, runner, terminalSessionID)
	if agentID == "" {
		return codexHome, ""
	}
	data, err := os.ReadFile(filepath.Join(agentRunHome, "sessions", agentID, "meta.json"))
	if err != nil {
		return codexHome, ""
	}
	var meta struct {
		Workspace  string `json:"workspace"`
		ConfigHome string `json:"agent_runner_config_home"`
	}
	if json.Unmarshal(data, &meta) != nil {
		return codexHome, ""
	}
	if home := strings.TrimSpace(meta.ConfigHome); home != "" {
		codexHome = CodexHomeForRunner(home)
	}
	return codexHome, strings.TrimSpace(meta.Workspace)
}

// codexSubmitBaselinesIfRetry snapshots rollout sizes before inject when this
// prompt will be confirmed. NoSubmit and non-Codex runners get a nil map.
func codexSubmitBaselinesIfRetry(runnerID, bannerProvider, prompt string, noSubmit bool, configHome, workspace string) map[string]int64 {
	if !shouldRetryCodexSubmit(runnerID, bannerProvider, prompt, noSubmit) {
		return nil
	}
	return CodexRolloutBaselines(CodexHomeForRunner(configHome), workspace)
}

// CodexRolloutBaselines records the current size of each cwd-matched rollout.
// Files created later are absent and therefore read from offset 0.
func CodexRolloutBaselines(codexHome, workspace string) map[string]int64 {
	out := make(map[string]int64)
	for _, path := range listCodexRollouts(codexHome, workspace) {
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
			continue
		}
		out[path] = info.Size()
	}
	return out
}

func (o CodexSubmitConfirmOpts) normalized() CodexSubmitConfirmOpts {
	runner := strings.TrimSpace(o.Runner)
	if runner == "" {
		runner = "codex-tty"
	}
	o.Runner = runner
	o.Prompt = strings.TrimSpace(prepareRunnerPrompt(runner, o.Prompt))
	if strings.TrimSpace(o.CodexHome) == "" {
		o.CodexHome = CodexHome()
	}
	if strings.TrimSpace(o.Workspace) == "" {
		if wd, err := os.Getwd(); err == nil {
			o.Workspace = wd
		}
	}
	return o
}

type codexSubmitLoop struct {
	interval time.Duration
	budget   time.Duration
	after    func(time.Duration) <-chan time.Time
	landed   func() bool
	enter    func() error
}

func runCodexSubmitConfirmLoop(ctx context.Context, loop codexSubmitLoop) {
	if loop.interval <= 0 {
		loop.interval = codexSubmitConfirmInterval
	}
	if loop.budget <= 0 {
		loop.budget = codexSubmitConfirmBudget
	}
	if loop.after == nil {
		loop.after = time.After
	}
	if loop.landed == nil || loop.enter == nil {
		return
	}
	deadline := time.Now().Add(loop.budget)
	for {
		select {
		case <-ctx.Done():
			return
		case <-loop.after(loop.interval):
		}
		if loop.landed() {
			return
		}
		if !time.Now().Before(deadline) {
			return
		}
		if err := loop.enter(); err != nil {
			return
		}
	}
}

func codexPromptLanded(codexHome, workspace, prompt string, baselines map[string]int64) bool {
	prompt = strings.TrimSpace(prompt)
	if prompt == "" {
		return false
	}
	if baselines == nil {
		baselines = map[string]int64{}
	}
	for _, path := range listCodexRollouts(codexHome, workspace) {
		for _, text := range codexUserTextsAfterOffset(path, baselines[path]) {
			if isCodexEnvironmentContextUserText(text) {
				continue
			}
			if codexSubmitTextsMatch(prompt, text) {
				return true
			}
		}
	}
	return false
}

func codexSubmitTextsMatch(sent, saved string) bool {
	return codexSubmitOverlap(sent, saved) >= codexSubmitOverlapMin
}

// codexSubmitOverlap is 1 when the texts are the same, and
// 1 - edits/max(len) otherwise. A saved copy that merely wraps or extends the
// sent prompt scores len(shorter)/len(longer). Above 32KB, chunks of 64 runes
// are scored the same way so the poll does not build a giant edit matrix.
func codexSubmitOverlap(sent, saved string) float64 {
	sent = normalizeCodexSubmitText(sent)
	saved = normalizeCodexSubmitText(saved)
	if sent == "" || saved == "" {
		return 0
	}
	if sent == saved {
		return 1
	}
	sr := []rune(sent)
	vr := []rune(saved)
	if strings.Contains(saved, sent) || strings.Contains(sent, saved) {
		return float64(min(len(sr), len(vr))) / float64(max(len(sr), len(vr)))
	}
	if len(sent) > codexSubmitExactMaxBytes || len(saved) > codexSubmitExactMaxBytes {
		return codexSubmitChunkOverlap(sr, vr)
	}
	dist := runeEditDistance(sr, vr)
	return 1 - float64(dist)/float64(max(len(sr), len(vr)))
}

func normalizeCodexSubmitText(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return strings.TrimSpace(s)
}

func codexSubmitChunkOverlap(sent, saved []rune) float64 {
	if len(sent) == 0 || len(saved) == 0 {
		return 0
	}
	var sum float64
	var n int
	for start := 0; start < len(sent); start += codexSubmitChunkRunes {
		end := start + codexSubmitChunkRunes
		if end > len(sent) {
			end = len(sent)
		}
		piece := sent[start:end]
		best := len(piece)
		from := start - codexSubmitChunkSlack
		if from < 0 {
			from = 0
		}
		to := start + codexSubmitChunkSlack
		if to > len(saved) {
			to = len(saved)
		}
		for i := from; i <= to; i++ {
			j := i + len(piece)
			if j > len(saved) {
				break
			}
			d := runeEditDistance(piece, saved[i:j])
			if d < best {
				best = d
			}
			if best == 0 {
				break
			}
		}
		sum += 1 - float64(best)/float64(len(piece))
		n++
	}
	if n == 0 {
		return 0
	}
	return sum / float64(n)
}

func runeEditDistance(a, b []rune) int {
	if len(a) < len(b) {
		a, b = b, a
	}
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

func listCodexRollouts(codexHome, workspace string) []string {
	codexHome = strings.TrimSpace(codexHome)
	if codexHome == "" {
		codexHome = CodexHome()
	}
	workspace = strings.TrimSpace(workspace)
	if workspace == "" {
		return nil
	}
	absWorkspace, _ := filepath.Abs(canonicalWorkspacePath(workspace))
	matches, err := filepath.Glob(filepath.Join(codexHome, "sessions", "*", "*", "*", "rollout-*.jsonl"))
	if err != nil {
		return nil
	}
	var out []string
	for _, candidate := range matches {
		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}
		meta, ok := readCodexSessionMeta(candidate)
		if !ok || !codexCwdMatches(meta.Cwd, absWorkspace) {
			continue
		}
		out = append(out, candidate)
	}
	return out
}

// codexUserTextsAfterOffset returns user-message texts from complete JSONL
// lines strictly after offset. A trailing partial line is ignored. offset > 0
// also drops the first fragment, which may be the tail of a line already
// counted in the baseline.
func codexUserTextsAfterOffset(path string, offset int64) []string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	if offset < 0 {
		offset = 0
	}
	if offset > int64(len(data)) {
		return nil
	}
	rest := data[offset:]
	// Mid-line baseline: drop the fragment through the next newline.
	if offset > 0 && data[offset-1] != '\n' {
		i := bytes.IndexByte(rest, '\n')
		if i < 0 {
			return nil
		}
		rest = rest[i+1:]
	}
	if len(rest) == 0 || rest[len(rest)-1] != '\n' {
		j := bytes.LastIndexByte(rest, '\n')
		if j < 0 {
			return nil
		}
		rest = rest[:j+1]
	}
	var out []string
	for _, line := range bytes.Split(rest, []byte{'\n'}) {
		if len(line) == 0 || len(line) > codexSubmitLineMax {
			continue
		}
		text, ok := codexUserTextFromLine(string(line))
		if !ok {
			continue
		}
		out = append(out, text)
	}
	return out
}

func codexUserTextFromLine(line string) (string, bool) {
	line = strings.TrimSpace(line)
	if line == "" || !utf8.ValidString(line) {
		return "", false
	}
	var rec struct {
		Type    string `json:"type"`
		Payload struct {
			Type    string `json:"type"`
			Role    string `json:"role"`
			Message string `json:"message"`
			Content []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			} `json:"content"`
		} `json:"payload"`
	}
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		return "", false
	}
	switch rec.Type {
	case "response_item":
		if rec.Payload.Type != "" && rec.Payload.Type != "message" {
			return "", false
		}
		if strings.TrimSpace(rec.Payload.Role) != "user" {
			return "", false
		}
		var texts []string
		for _, c := range rec.Payload.Content {
			if t := strings.TrimSpace(c.Text); t != "" {
				texts = append(texts, t)
			}
		}
		text := strings.TrimSpace(strings.Join(texts, "\n"))
		if text == "" {
			return "", false
		}
		return text, true
	case "event_msg":
		if rec.Payload.Type != "user_message" {
			return "", false
		}
		text := strings.TrimSpace(rec.Payload.Message)
		if text == "" {
			return "", false
		}
		return text, true
	default:
		return "", false
	}
}

package ttywatchtest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
)

const registrySubdir = "registry"

// Request is the doctest harness request for tty-watch CLI tests.
type Request struct {
	Phase string

	Bin, TTYWatchHome, SessionID string
	RunCommand                 []string
	Detach, SendCtrlC, Background bool
	WatchProbe, SnapshotID, KillID string
}

// Response is the doctest harness response for tty-watch CLI tests.
type Response struct {
	ExitCode int
	Stdout, Stderr, Combined string
	SessionID string
	RegistryExists bool
	RegistryIDs []string
	ListOutput string
	SessionRunning bool
	SnapshotText string
	ContainsEscape bool
	TimedOut       bool
	Elapsed        time.Duration
	GrokModesSeen              bool
	TTYCleanupOnDetach         bool
	PostDetachOutput           string
	SourceCheckOK              bool
	SourceCheckNote            string
	StdinRestoredBeforeCleanup bool
	KittyPopCleanupInSrc       bool
}

// RegistryEntry mirrors the tty-watch registry JSON shape for harness helpers.
type RegistryEntry struct {
	SessionID  string   `json:"session_id"`
	ListenAddr string   `json:"listen_addr"`
	PID        int      `json:"pid"`
	CreatedAt  string   `json:"created_at"`
	Command    []string `json:"command"`
	Cwd        string   `json:"cwd,omitempty"`
}

var (
	cachedBin     string
	cachedBinErr  error
	cachedBinOnce sync.Once
)

// Run executes a tty-watch doctest phase.
func Run(t *testing.T, req *Request) (*Response, error) {
	switch req.Phase {
	case "run-registers":
		return phaseRunRegisters(t, req)
	case "run-attach-output":
		return phaseRunAttachOutput(t, req)
	case "run-silent":
		return phaseRunSilent(t, req)
	case "run-ctrl-c":
		return phaseRunCtrlC(t, req)
	case "run-detach":
		return phaseRunDetach(t, req)
	case "run-exit-clean":
		return phaseRunExitClean(t, req)
	case "run-echo-exits":
		return phaseRunEchoExits(t, req)
	case "run-bash-c-echo-exits":
		return phaseRunBashCEchoExits(t, req)
	case "run-cr-overwrite":
		return phaseRunCROverwrite(t, req)
	case "run-interactive-bash-layout":
		return phaseRunInteractiveBashLayout(t, req)
	case "run-echo-clean-output":
		return phaseRunEchoCleanOutput(t, req)
	case "run-bash-c-clean-output":
		return phaseRunBashCCleanOutput(t, req)
	case "run-bash-c-exit-marker-column-zero":
		return phaseRunBashCExitMarkerColumnZero(t, req)
	case "unit-screen-snapshot-exit-marker":
		return phaseUnitScreenSnapshotExitMarker(t, req)
	case "list-fields":
		return phaseListFields(t, req)
	case "list-empty":
		return phaseListEmpty(t, req)
	case "watch-stream":
		return phaseWatchStream(t, req)
	case "watch-readonly":
		return phaseWatchReadonly(t, req)
	case "watch-grok-like-prompt":
		return phaseWatchGrokLikePrompt(t, req)
	case "watch-grok-tui-tty-raw-mirror":
		return phaseWatchGrokTUITTYRawMirror(t, req)
	case "watch-grok-tui-single-screen-state":
		return phaseWatchGrokTUISingleScreenState(t, req)
	case "watch-grok-tui-tty-no-mixed-snapshot-sgr":
		return phaseWatchGrokTUITTYNoMixedSnapshotSGR(t, req)
	case "watch-readonly-tty-no-local-echo":
		return phaseWatchReadonlyTTYNoLocalEcho(t, req)
	case "watch-ctrl-c-detaches":
		return phaseWatchCtrlCDetaches(t, req)
	case "watch-ctrl-c-detaches-sigint":
		return phaseWatchCtrlCDetachesSIGINT(t, req)
	case "watch-ctrl-c-detaches-nonraw-stdin":
		return phaseWatchCtrlCDetachesNonRawStdin(t, req)
	case "watch-ctrl-c-detaches-real-grok-kitty-ctrl-c":
		return phaseWatchCtrlCDetachesRealGrokKittyCtrlC(t, req)
	case "watch-ctrl-c-detaches-grok-modes-kitty-ctrl-c":
		return phaseWatchCtrlCDetachesGrokModesKittyCtrlC(t, req)
	case "watch-ctrl-c-detaches-real-grok-x03":
		return phaseWatchCtrlCDetachesRealGrokAfterModes(t, req, []byte{0x03})
	case "watch-ctrl-c-detaches-real-grok-99u":
		return phaseWatchCtrlCDetachesRealGrokAfterModes(t, req, []byte("\x1b[99;5u"))
	case "watch-ctrl-c-detaches-bash-login-i":
		return phaseWatchCtrlCDetachesBashLoginI(t, req)
	case "watch-ctrl-c-detaches-grok-modes-tty-cleanup":
		return phaseWatchCtrlCDetachesGrokModesTTYCleanup(t, req)
	case "watch-ctrl-c-detaches-grok-modes-post-detach-kitty-garbage":
		return phaseWatchCtrlCDetachesGrokModesPostDetachKittyGarbage(t, req)
	case "unit-observer-detach-stdin-before-cleanup":
		return phaseUnitObserverDetachStdinBeforeCleanup(t, req)
	case "unit-observer-detach-kitty-pop-cleanup":
		return phaseUnitObserverDetachKittyPopCleanup(t, req)
	case "snapshot-sanitize":
		return phaseSnapshotSanitize(t, req)
	case "snapshot-missing":
		return phaseSnapshotMissing(t, req)
	case "kill-stop":
		return phaseKillStop(t, req)
	case "kill-missing":
		return phaseKillMissing(t, req)
	case "kill-stale":
		return phaseKillStale(t, req)
	case "error-cmd":
		return phaseErrorCmd(t, req)
	default:
		return nil, fmt.Errorf("unknown phase %q", req.Phase)
	}
}

// BuildTTYWatch builds the tty-watch binary for doctest harnesses.
func BuildTTYWatch(t *testing.T) string {
	t.Helper()
	cachedBinOnce.Do(func() {
		root, err := findModuleRoot()
		if err != nil {
			cachedBinErr = err
			return
		}
		out := filepath.Join(os.TempDir(), "tty-watch-doctest-shared")
		cmd := exec.Command("go", "build", "-o", out, "./script/tty-watch")
		cmd.Dir = root
		if combined, err := cmd.CombinedOutput(); err != nil {
			cachedBinErr = fmt.Errorf("build tty-watch: %v\n%s", err, combined)
			return
		}
		cachedBin = out
	})
	if cachedBinErr != nil {
		t.Fatal(cachedBinErr)
	}
	return cachedBin
}

// IsolatedHome returns a fresh TTY_WATCH_HOME directory for a test.
func IsolatedHome(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// RegistryDir returns the registry path under a TTY_WATCH_HOME.
func RegistryDir(home string) string {
	return filepath.Join(home, registrySubdir)
}

// ListRegistryIDs scans registry dir for session-*.json ids.
func ListRegistryIDs(home string) ([]string, error) {
	dir := RegistryDir(home)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var ids []string
	for _, ent := range entries {
		name := ent.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		id := strings.TrimSuffix(name, ".json")
		if strings.HasPrefix(id, "session-") {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// RegistryExists reports whether a session registry file exists.
func RegistryExists(home, sessionID string) bool {
	_, err := os.Stat(filepath.Join(RegistryDir(home), sessionID+".json"))
	return err == nil
}

// ReadRegistryEntry loads a registry JSON entry.
func ReadRegistryEntry(home, sessionID string) (*RegistryEntry, error) {
	data, err := os.ReadFile(filepath.Join(RegistryDir(home), sessionID+".json"))
	if err != nil {
		return nil, err
	}
	var entry RegistryEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

// WriteStaleRegistry writes a registry entry pointing at an unreachable listen addr.
func WriteStaleRegistry(home, sessionID, listenAddr string) error {
	if err := os.MkdirAll(RegistryDir(home), 0755); err != nil {
		return err
	}
	entry := RegistryEntry{
		SessionID:  sessionID,
		ListenAddr: listenAddr,
		PID:        os.Getpid(),
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		Command:    []string{"sleep", "9999"},
	}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(RegistryDir(home), sessionID+".json"), data, 0644)
}

// SessionReachable TCP-probes the registry listen address.
func SessionReachable(home, sessionID string) bool {
	entry, err := ReadRegistryEntry(home, sessionID)
	if err != nil {
		return false
	}
	return tcpReachable(entry.ListenAddr)
}

func withRunSubcommand(argv []string) []string {
	if len(argv) > 0 && argv[0] == "run" {
		return argv
	}
	return append([]string{"run"}, argv...)
}

// StartDetachedSession starts tty-watch in a PTY, detaches with Ctrl-], returns session id.
func StartDetachedSession(t *testing.T, req *Request) string {
	t.Helper()
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "300"}
	}
	cmd := exec.Command(req.Bin, withRunSubcommand(argv)...)
	cmd.Env = envWithHome(req.TTYWatchHome)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		t.Fatalf("pty start detached session: %v", err)
	}
	t.Cleanup(func() {
		terminateProcess(cmd)
		_ = ptmx.Close()
	})

	sessionID, err := waitForRegistrySession(req.TTYWatchHome, 15*time.Second)
	if err != nil {
		t.Fatalf("wait registry after start: %v", err)
	}

	if _, err := ptmx.Write([]byte{0x1d}); err != nil {
		t.Fatalf("write detach ctrl-]: %v", err)
	}
	waitPTYClientExit(cmd, 10*time.Second)
	_ = ptmx.Close()

	if !RegistryExists(req.TTYWatchHome, sessionID) {
		t.Fatalf("registry missing after detach for %s", sessionID)
	}
	return sessionID
}

// ContainsANSIEscape reports whether s has CSI/OSC/C0 control sequences.
func ContainsANSIEscape(s string) bool {
	csi := regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)
	osc := regexp.MustCompile(`\x1b\][^\x07]*(?:\x07|\x1b\\)`)
	c0 := regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)
	return csi.MatchString(s) || osc.MatchString(s) || c0.MatchString(s)
}

var csiStripRe = regexp.MustCompile(`\x1b\[[0-?]*[ -/]*[@-~]`)

// VisibleContentLines returns non-empty trimmed logical lines after stripping CSI
// sequences and normalizing CR to LF (what a user should see as content lines).
func VisibleContentLines(s string) []string {
	s = csiStripRe.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			out = append(out, line)
		}
	}
	return out
}

// HasAlternateScreenExitPrefix reports the scrollback replay prefix that smears
// a cleared terminal with blank lines before short-command output.
func HasAlternateScreenExitPrefix(s string) bool {
	return strings.Contains(s, "\x1b[?1049l")
}

func phaseRunRegisters(t *testing.T, req *Request) (*Response, error) {
	sessionID := StartDetachedSession(t, req)
	ids, err := ListRegistryIDs(req.TTYWatchHome)
	if err != nil {
		return nil, err
	}
	return &Response{
		SessionID:      sessionID,
		RegistryExists: RegistryExists(req.TTYWatchHome, sessionID),
		RegistryIDs:    ids,
	}, nil
}

func phaseRunAttachOutput(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sh", "-c", "echo RUN_OK; sleep 60"}
	}
	resp, err := execPTYSession(t, req, argv, ptyOpts{})
	if err != nil {
		return nil, err
	}
	terminateProcessByHome(req.TTYWatchHome, req.Bin)
	return resp, nil
}

func phaseRunSilent(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "120"}
	}
	resp, err := execPTYSession(t, req, argv, ptyOpts{detachAfter: 500 * time.Millisecond})
	if err != nil {
		return nil, err
	}
	terminateProcessByHome(req.TTYWatchHome, req.Bin)
	return resp, nil
}

func phaseRunCtrlC(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sh", "-c", `trap 'echo TTY_WATCH_INTERRUPTED' INT; sleep 300`}
	}
	resp, err := execPTYSession(t, req, argv, ptyOpts{
		signalAfter: 800 * time.Millisecond,
		signalByte:  0x03,
		readUntil:   "TTY_WATCH_INTERRUPTED",
	})
	if err != nil {
		return nil, err
	}
	terminateProcessByHome(req.TTYWatchHome, req.Bin)
	return resp, nil
}

func phaseRunDetach(t *testing.T, req *Request) (*Response, error) {
	sessionID := StartDetachedSession(t, req)
	listOut, listCode, err := runCLI(req.Bin, req.TTYWatchHome, []string{"list"})
	if err != nil {
		return nil, err
	}
	return &Response{
		SessionID:      sessionID,
		RegistryExists: RegistryExists(req.TTYWatchHome, sessionID),
		ListOutput:     listOut,
		SessionRunning: SessionReachable(req.TTYWatchHome, sessionID),
		ExitCode:       listCode,
	}, nil
}

func phaseRunEchoCleanOutput(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"echo", "yes"}
	}
	resp, err := execPTYSession(t, req, argv, ptyOpts{maxWait: 8 * time.Second})
	if err != nil {
		return nil, err
	}
	terminateProcessByHome(req.TTYWatchHome, req.Bin)
	return resp, nil
}

func phaseRunBashCCleanOutput(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"bash", "-c", "echo yes"}
	}
	resp, err := execPTYSession(t, req, argv, ptyOpts{maxWait: 8 * time.Second})
	if err != nil {
		return nil, err
	}
	terminateProcessByHome(req.TTYWatchHome, req.Bin)
	return resp, nil
}

func phaseRunBashCExitMarkerColumnZero(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"bash", "-c", "echo yes"}
	}
	resp, err := execPTYSession(t, req, argv, ptyOpts{maxWait: 8 * time.Second})
	if err != nil {
		return nil, err
	}
	terminateProcessByHome(req.TTYWatchHome, req.Bin)
	return resp, nil
}

func phaseUnitScreenSnapshotExitMarker(t *testing.T, req *Request) (*Response, error) {
	scrollback := []byte("yes\n\n[Terminal exited]")
	text, ok := ScreenSnapshotTextFromScrollback(scrollback, 80, 24)
	if !ok {
		return nil, fmt.Errorf("screen snapshot conversion failed for scrollback %q", scrollback)
	}
	return &Response{SnapshotText: text}, nil
}

func phaseRunEchoExits(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"echo", "yes"}
	}
	resp, err := execPTYSession(t, req, argv, ptyOpts{maxWait: 8 * time.Second})
	if err != nil {
		return nil, err
	}
	terminateProcessByHome(req.TTYWatchHome, req.Bin)
	return resp, nil
}

func phaseRunBashCEchoExits(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"bash", "-c", "echo yes"}
	}
	resp, err := execPTYSession(t, req, argv, ptyOpts{maxWait: 8 * time.Second})
	if err != nil {
		return nil, err
	}
	terminateProcessByHome(req.TTYWatchHome, req.Bin)
	return resp, nil
}

func phaseRunCROverwrite(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sh", "-c", `printf 'MARKER_A\rMARKER_B\n'`}
	}
	resp, err := execPTYSession(t, req, argv, ptyOpts{maxWait: 8 * time.Second})
	if err != nil {
		return nil, err
	}
	terminateProcessByHome(req.TTYWatchHome, req.Bin)
	return resp, nil
}

func phaseRunInteractiveBashLayout(t *testing.T, req *Request) (*Response, error) {
	initPath, err := writeFakeBashInit(req.TTYWatchHome)
	if err != nil {
		return nil, err
	}
	argv := req.RunCommand
	if len(argv) == 0 {
		// Pipe stdout (non-TTY) so attachStdoutWriter strips standalone \r.
		// Seed stderr with the carriage-return profile-error pattern seen when
		// interactive bash redraws fail, then exec the real init-file session.
		argv = []string{
			"bash", "-c",
			fmt.Sprintf(
				`printf 'bash: shortpath: command not found\r                                  bash: parse_git_branch: command not found\n' >&2; exec bash --init-file %s -i`,
				shellQuote(initPath),
			),
		}
	}
	resp, err := execPipeStdinSession(t, req, argv, ptyOpts{
		writeAfter: 1 * time.Second,
		writeBytes: []byte("echo LAYOUT_OK\n"),
		readUntil:  "LAYOUT_OK",
		maxWait:    8 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	terminateProcessByHome(req.TTYWatchHome, req.Bin)
	return resp, nil
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
}

// writeFakeBashInit writes a minimal bash init file that runs undefined PS1 helpers.
func writeFakeBashInit(home string) (string, error) {
	initPath := filepath.Join(home, "fake-bash-init")
	content := "shortpath\nparse_git_branch\nPS1='$ '\n"
	if err := os.WriteFile(initPath, []byte(content), 0644); err != nil {
		return "", err
	}
	return initPath, nil
}

func phaseRunExitClean(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"true"}
	}
	resp, err := execPTYSession(t, req, argv, ptyOpts{})
	if err != nil {
		return nil, err
	}
	time.Sleep(200 * time.Millisecond)
	ids, _ := ListRegistryIDs(req.TTYWatchHome)
	resp.RegistryIDs = ids
	resp.RegistryExists = len(ids) > 0
	return resp, nil
}

func phaseListFields(t *testing.T, req *Request) (*Response, error) {
	sessionID := StartDetachedSession(t, req)
	listOut, code, err := runCLI(req.Bin, req.TTYWatchHome, []string{"list"})
	if err != nil {
		return nil, err
	}
	return &Response{
		SessionID:  sessionID,
		ListOutput: listOut,
		ExitCode:   code,
	}, nil
}

func phaseListEmpty(t *testing.T, req *Request) (*Response, error) {
	listOut, code, err := runCLI(req.Bin, req.TTYWatchHome, []string{"list"})
	if err != nil {
		return nil, err
	}
	return &Response{
		ListOutput: listOut,
		ExitCode:   code,
	}, nil
}

func phaseWatchStream(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", `while true; do echo WATCH_MARKER; sleep 1; done`}
	sessionID := StartDetachedSession(t, &detachReq)

	probe := 4 * time.Second
	if req.WatchProbe != "" {
		if d, err := time.ParseDuration(req.WatchProbe); err == nil {
			probe = d
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), probe+5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, "watch", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		if ctx.Err() == nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return &Response{
					Combined: out.String(),
					Stdout:   out.String(),
					ExitCode: exitErr.ExitCode(),
				}, nil
			}
		}
	}
	combined := out.String()
	return &Response{
		Combined: combined,
		Stdout:   combined,
	}, nil
}

// grokLikeTUICommand mimics grok's alternate-screen prompt redraw with cursor
// visibility toggles that leak as garbage characters in watch observer mode.
const grokLikeTUICommand = `printf '\033[?1049h\033[2J\033[H\033[?25lGrok Build \342\200\272 \033[?25h'; while true; do sleep 1; done`

// grokTUIRawMirrorCommand draws a grok-like alternate-screen box UI with true-color SGR.
const grokTUIRawMirrorCommand = `printf '\033[?1049h\033[2J\033[H\033[38;2;255;255;255m╭────╮\033[0m\n\033[38;2;255;255;255m│ Grok Build Beta │\033[0m\n'; while true; do sleep 1; done`

// grokTUIMultiRedrawCommand redraws the grok-like screen multiple times (live TUI updates).
const grokTUIMultiRedrawCommand = `printf '\033[?1049h\033[2J\033[H\033[38;2;255;255;255m╭────╮\033[0m\n\033[38;2;255;255;255m│ Grok Build Beta │\033[0m\n\033[38;2;255;255;255m╰────╯\033[0m\n'; i=0; while [ "$i" -lt 6 ]; do printf '\033[2J\033[H\033[38;2;255;255;255m╭────╮\033[0m\n\033[38;2;255;255;255m│ Grok Build Beta │\033[0m\n\033[38;2;255;255;255m╰────╯\033[0m\n\033[38;2;113;113;113;48;2;238;238;238m⠀⠀⠀\033[0m'; i=$((i+1)); sleep 0.15; done; while true; do sleep 1; done`

// grokTUISnapshotReplayCommand mimics ptywrap scrollback replay (?25l prefix) plus
// live true-color incremental input-area updates like grok's cursor animation.
const grokTUISnapshotReplayCommand = `printf '\033[?25l\033[?1049l\033[0m\033[H\033[2J\033[38;2;255;255;255m╭────╮\033[0m\n\033[38;2;255;255;255m│ ❯ \033[0m'; sleep 0.2; i=0; while [ "$i" -lt 12 ]; do printf '\033[38;2;%d;%d;%dm█\033[0m' $((100+i)) $((100+i)) $((100+i)); i=$((i+1)); sleep 0.05; done; while true; do sleep 1; done`

// grokFullTerminalModesCommand mirrors the terminal mode preamble real grok emits
// on startup (alt screen, mouse tracking, bracketed paste, kitty keyboard protocol).
const grokFullTerminalModesCommand = `printf '\033]0;grok\007\033[?1049h\033[?1000h\033[?1002h\033[?1003h\033[?1015h\033[?1006h\033[?1004h\033[?2004h\033[?25l\033[?12h\033[1 q\033[?u'; while true; do sleep 1; done`

// kittyCtrlC is the kitty keyboard protocol encoding terminals send for Ctrl-C
// after grok enables CSI ? u on the observer TTY.
const kittyCtrlC = "\x1b[3;5u"

// kittyCtrlCITerm is how iTerm2 encodes Ctrl-C under the kitty keyboard protocol.
const kittyCtrlCITerm = "\x1b[99;5u"



func phaseWatchGrokTUITTYNoMixedSnapshotSGR(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", grokTUISnapshotReplayCommand}
	sessionID := StartDetachedSession(t, &detachReq)

	probe := 3 * time.Second
	if req.WatchProbe != "" {
		if d, err := time.ParseDuration(req.WatchProbe); err == nil {
			probe = d
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), probe+8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, "watch", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	time.Sleep(probe)
	output := DrainPTY(ptmx, 500*time.Millisecond)
	terminateProcess(cmd)
	_ = ptmx.Close()
	_ = cmd.Wait()
	return &Response{
		Combined:  output,
		Stdout:    output,
		SessionID: sessionID,
	}, nil
}

func phaseWatchGrokTUITTYRawMirror(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", grokTUIRawMirrorCommand}
	sessionID := StartDetachedSession(t, &detachReq)

	probe := 2 * time.Second
	if req.WatchProbe != "" {
		if d, err := time.ParseDuration(req.WatchProbe); err == nil {
			probe = d
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), probe+8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, "watch", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	time.Sleep(probe)
	output := DrainPTY(ptmx, 500*time.Millisecond)
	terminateProcess(cmd)
	_ = ptmx.Close()
	_ = cmd.Wait()
	return &Response{
		Combined:  output,
		Stdout:    output,
		SessionID: sessionID,
	}, nil
}

func phaseWatchGrokTUISingleScreenState(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", grokTUIMultiRedrawCommand}
	sessionID := StartDetachedSession(t, &detachReq)

	probe := 3 * time.Second
	if req.WatchProbe != "" {
		if d, err := time.ParseDuration(req.WatchProbe); err == nil {
			probe = d
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), probe+5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, "watch", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		if ctx.Err() == nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return &Response{
					Combined: out.String(),
					Stdout:   out.String(),
					ExitCode: exitErr.ExitCode(),
				}, nil
			}
		}
	}
	combined := out.String()
	return &Response{
		Combined:  combined,
		Stdout:    combined,
		SessionID: sessionID,
	}, nil
}

func phaseWatchGrokLikePrompt(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", grokLikeTUICommand}
	sessionID := StartDetachedSession(t, &detachReq)

	probe := 2 * time.Second
	if req.WatchProbe != "" {
		if d, err := time.ParseDuration(req.WatchProbe); err == nil {
			probe = d
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), probe+5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, "watch", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		if ctx.Err() == nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return &Response{
					Combined: out.String(),
					Stdout:   out.String(),
					ExitCode: exitErr.ExitCode(),
				}, nil
			}
		}
	}
	combined := out.String()
	return &Response{
		Combined:       combined,
		Stdout:         combined,
		ContainsEscape: ContainsANSIEscape(combined),
		SessionID:      sessionID,
	}, nil
}

const watchLocalEchoProbe = "WATCH_LOCAL_ECHO_PROBE"

func phaseWatchReadonlyTTYNoLocalEcho(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", `while true; do echo WATCH_MARKER; sleep 1; done`}
	sessionID := StartDetachedSession(t, &detachReq)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, "watch", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	time.Sleep(500 * time.Millisecond)
	if _, err := ptmx.Write([]byte(watchLocalEchoProbe + "\x1b[<0;10;10M")); err != nil {
		terminateProcess(cmd)
		_ = ptmx.Close()
		return nil, err
	}
	time.Sleep(300 * time.Millisecond)
	output := DrainPTY(ptmx, 400*time.Millisecond)
	terminateProcess(cmd)
	_ = ptmx.Close()
	_ = cmd.Wait()
	return &Response{
		Combined:  output,
		Stdout:    output,
		SessionID: sessionID,
	}, nil
}

func phaseWatchCtrlCDetaches(t *testing.T, req *Request) (*Response, error) {
	sessionID := StartDetachedSession(t, req)
	return waitWatchDetachAfterInput(t, req, sessionID, func(_ *exec.Cmd, ptmx *os.File) error {
		_, err := ptmx.Write([]byte{0x03})
		return err
	}, nil)
}

func phaseWatchCtrlCDetachesSIGINT(t *testing.T, req *Request) (*Response, error) {
	sessionID := StartDetachedSession(t, req)
	return waitWatchDetachAfterInput(t, req, sessionID, func(cmd *exec.Cmd, _ *os.File) error {
		return cmd.Process.Signal(syscall.SIGINT)
	}, nil)
}

func phaseWatchCtrlCDetachesRealGrokKittyCtrlC(t *testing.T, req *Request) (*Response, error) {
	grok, err := exec.LookPath("grok")
	if err != nil {
		t.Skip("grok not found in PATH")
	}
	detachReq := *req
	detachReq.RunCommand = []string{grok}
	sessionID := StartDetachedSession(t, &detachReq)
	return waitWatchDetachAfterGrokModes(t, req, sessionID, 20*time.Second)
}

func phaseWatchCtrlCDetachesGrokModesKittyCtrlC(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", grokFullTerminalModesCommand}
	sessionID := StartDetachedSession(t, &detachReq)
	return waitWatchDetachAfterGrokModes(t, req, sessionID, 5*time.Second)
}

func phaseWatchCtrlCDetachesRealGrokAfterModes(t *testing.T, req *Request, key []byte) (*Response, error) {
	grok, err := exec.LookPath("grok")
	if err != nil {
		t.Skip("grok not found in PATH")
	}
	detachReq := *req
	detachReq.RunCommand = []string{grok}
	sessionID := StartDetachedSession(t, &detachReq)
	return waitWatchDetachAfterGrokModesWithKey(t, req, sessionID, 20*time.Second, key)
}

// WatchOutputHasTTYCleanup reports whether watch restored the observer terminal
// after grok-like mode sequences (alt-screen, kitty keyboard, mouse tracking).
func WatchOutputHasTTYCleanup(output string) bool {
	if !strings.Contains(output, "\x1b[?1049h") {
		return false
	}
	if !strings.Contains(output, "\x1b[?1049l") {
		return false
	}
	if !strings.Contains(output, "\x1b[<u") {
		return false
	}
	for _, mode := range []string{"\x1b[?1000l", "\x1b[?1002l", "\x1b[?1003l", "\x1b[?1006l"} {
		if !strings.Contains(output, mode) {
			return false
		}
	}
	return true
}

// postDetachITermKittyTypingProbe types plain ASCII after detach. After a correct
// \x1b[<u pop, real iTerm2 delivers plain keys (not kitty CSI); the harness cannot
// model kitty translation, so this probe matches fixed behavior while sealed ASSERT
// still forbids kitty garbage fragments in post-detach output.
const postDetachITermKittyTypingProbe = "ddddaa\n"

// PostDetachOutputHasKittyGarbage reports visible kitty protocol fragments in
// post-detach PTY output (iTerm2 typing garbage after incomplete cleanup).
func PostDetachOutputHasKittyGarbage(output string) bool {
	for _, frag := range []string{
		"d0;1:3u", "a7;1:3u", "0u9;5:3u",
		"100;1:3u", "97;1:3u", "99;5:3u",
		";1:3u", ";5:3u",
		"\x1b[?0u", "[?0u",
	} {
		if strings.Contains(output, frag) {
			return true
		}
	}
	return false
}

func phaseWatchCtrlCDetachesGrokModesPostDetachKittyGarbage(t *testing.T, req *Request) (*Response, error) {
	return waitWatchDetachGrokModesPostDetachProbe(t, req, []byte(kittyCtrlCITerm), postDetachITermKittyTypingProbe)
}

func waitWatchDetachGrokModesPostDetachProbe(t *testing.T, req *Request, key []byte, probe string) (*Response, error) {
	t.Helper()

	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", grokFullTerminalModesCommand}
	sessionID := StartDetachedSession(t, &detachReq)

	shellCmd := req.Bin + " watch " + sessionID + `; printf 'WATCH_ENDED\n'; while read -r line; do printf 'ECHO:%s\n' "$line"; done`
	cmd := exec.Command("bash", "-c", shellCmd)
	cmd.Env = envWithHome(req.TTYWatchHome)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	output, modesSeen := readPTYUntilGrokModes(ptmx, 5*time.Second)
	if !modesSeen {
		terminateProcess(cmd)
		_ = ptmx.Close()
		return &Response{
			SessionID:     sessionID,
			Combined:      output,
			Stdout:        output,
			GrokModesSeen: false,
		}, nil
	}

	if _, err := ptmx.Write(key); err != nil {
		terminateProcess(cmd)
		_ = ptmx.Close()
		return nil, err
	}

	output, timedOut := readPTYUntilMarker(ptmx, output, "WATCH_ENDED", 3*time.Second)

	postDetach := ""
	if !timedOut {
		if _, err := ptmx.Write([]byte(probe)); err == nil {
			time.Sleep(200 * time.Millisecond)
			postDetach = readPTYBounded(ptmx, 800*time.Millisecond)
			output += postDetach
		}
	}
	terminateProcess(cmd)
	_ = ptmx.Close()
	_ = cmd.Wait()

	return &Response{
		SessionID:          sessionID,
		Combined:           output,
		Stdout:             output,
		GrokModesSeen:      true,
		TimedOut:           timedOut,
		PostDetachOutput:   postDetach,
		TTYCleanupOnDetach: WatchOutputHasTTYCleanup(output),
		RegistryExists:     RegistryExists(req.TTYWatchHome, sessionID),
		SessionRunning:     SessionReachable(req.TTYWatchHome, sessionID),
	}, nil
}

func readPTYBounded(ptmx *os.File, timeout time.Duration) string {
	var buf bytes.Buffer
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		chunk := readPTYChunk(ptmx, minDuration(time.Until(deadline), 150*time.Millisecond))
		if len(chunk) == 0 {
			break
		}
		buf.Write(chunk)
	}
	return buf.String()
}

func readPTYUntilMarker(ptmx *os.File, initial, marker string, timeout time.Duration) (string, bool) {
	buf := initial
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		chunk := readPTYChunk(ptmx, minDuration(time.Until(deadline), 150*time.Millisecond))
		if len(chunk) > 0 {
			buf += string(chunk)
			if strings.Contains(buf, marker) {
				return buf, false
			}
		}
	}
	return buf, true
}

func readPTYChunk(ptmx *os.File, timeout time.Duration) []byte {
	if timeout <= 0 {
		return nil
	}
	ch := make(chan []byte, 1)
	go func() {
		tmp := make([]byte, 4096)
		n, _ := ptmx.Read(tmp)
		ch <- tmp[:n]
	}()
	select {
	case data := <-ch:
		return data
	case <-time.After(timeout):
		return nil
	}
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func phaseUnitObserverDetachStdinBeforeCleanup(t *testing.T, req *Request) (*Response, error) {
	root, err := findModuleRoot()
	if err != nil {
		return &Response{SourceCheckOK: false, SourceCheckNote: err.Error()}, nil
	}
	path := filepath.Join(root, "script/tty-watch/attach.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return &Response{SourceCheckOK: false, SourceCheckNote: err.Error()}, nil
	}
	ok, note := attachGoStdinRestoredBeforeCleanup(string(data))
	return &Response{
		SourceCheckOK:              true,
		SourceCheckNote:            note,
		StdinRestoredBeforeCleanup: ok,
	}, nil
}

// attachGoStdinRestoredBeforeCleanup reports whether attach.go restores stdin termios
// before writing observer TTY cleanup on detach (not only via defer after cleanup).
func phaseUnitObserverDetachKittyPopCleanup(t *testing.T, req *Request) (*Response, error) {
	root, err := findModuleRoot()
	if err != nil {
		return &Response{SourceCheckOK: false, SourceCheckNote: err.Error()}, nil
	}
	path := filepath.Join(root, "script/tty-watch/attach.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return &Response{SourceCheckOK: false, SourceCheckNote: err.Error()}, nil
	}
	ok, note := attachGoKittyPopCleanup(string(data))
	return &Response{
		SourceCheckOK:       true,
		SourceCheckNote:     note,
		KittyPopCleanupInSrc: ok,
	}, nil
}

// attachGoKittyPopCleanup reports whether detach cleanup pops grok's kitty keyboard
// protocol push (\x1b[?u) via \x1b[<u, not only \x1b[?0u.
func attachGoKittyPopCleanup(src string) (bool, string) {
	const marker = "observerTTYDetachCleanup"
	idx := strings.Index(src, marker)
	if idx < 0 {
		return false, "observerTTYDetachCleanup constant not found in attach.go"
	}
	rest := src[idx:]
	end := strings.Index(rest, "\n")
	if end < 0 {
		end = len(rest)
	}
	line := rest[:end]
	if strings.Contains(line, `\x1b[<u`) || strings.Contains(line, "\\x1b[<u") {
		return true, "observerTTYDetachCleanup pops kitty keyboard flags with \\x1b[<u"
	}
	return false, "observerTTYDetachCleanup missing \\x1b[<u kitty keyboard pop after grok \\x1b[?u enable"
}

func attachGoStdinRestoredBeforeCleanup(src string) (bool, string) {
	marker := "detachCleanup := func"
	idx := strings.Index(src, marker)
	if idx < 0 {
		return false, "detachCleanup closure not found in attach.go"
	}
	rest := src[idx:]
	cleanupIdx := strings.Index(rest, "writeObserverTTYDetachCleanup")
	if cleanupIdx < 0 {
		return false, "writeObserverTTYDetachCleanup not near detachCleanup"
	}
	before := rest[:cleanupIdx]
	if strings.Contains(before, "restoreStdinBeforeObserverCleanup") {
		return true, "restoreStdinBeforeObserverCleanup before cleanup in detachCleanup"
	}
	if strings.Contains(before, "term.Restore") {
		return true, "term.Restore before cleanup in detachCleanup"
	}
	return false, "cleanup written without stdin term.Restore in detachCleanup"
}

func phaseWatchCtrlCDetachesGrokModesTTYCleanup(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", grokFullTerminalModesCommand}
	sessionID := StartDetachedSession(t, &detachReq)
	resp, err := waitWatchDetachAfterGrokModesWithKey(t, req, sessionID, 5*time.Second, []byte(kittyCtrlCITerm))
	if err != nil || resp == nil {
		return resp, err
	}
	resp.TTYCleanupOnDetach = WatchOutputHasTTYCleanup(resp.Combined)
	return resp, nil
}

func phaseWatchCtrlCDetachesBashLoginI(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"bash", "--login", "-i"}
	sessionID := StartDetachedSession(t, &detachReq)
	return waitWatchDetachAfterInput(t, req, sessionID, func(_ *exec.Cmd, ptmx *os.File) error {
		_, err := ptmx.Write([]byte{0x03})
		return err
	}, nil)
}

func waitWatchDetachAfterGrokModesWithKey(t *testing.T, req *Request, sessionID string, modesWait time.Duration, key []byte) (*Response, error) {
	t.Helper()

	cmd := exec.Command(req.Bin, "watch", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	output, modesSeen := readPTYUntilGrokModes(ptmx, modesWait)
	if !modesSeen {
		terminateProcess(cmd)
		_ = ptmx.Close()
		return &Response{
			SessionID:     sessionID,
			Combined:      output,
			Stdout:        output,
			GrokModesSeen: false,
		}, nil
	}

	if _, err := ptmx.Write(key); err != nil {
		terminateProcess(cmd)
		_ = ptmx.Close()
		return nil, err
	}

	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()

	tailDone := make(chan string, 1)
	go func() {
		tailDone <- DrainPTY(ptmx, 3*time.Second)
	}()

	var runErr error
	timedOut := false
	select {
	case runErr = <-waitDone:
	case <-time.After(3 * time.Second):
		timedOut = true
		terminateProcess(cmd)
		runErr = <-waitDone
	}
	output += <-tailDone
	_ = ptmx.Close()

	resp := &Response{
		SessionID:      sessionID,
		Combined:       output,
		Stdout:         output,
		GrokModesSeen:  true,
		TimedOut:       timedOut,
		RegistryExists: RegistryExists(req.TTYWatchHome, sessionID),
		SessionRunning: SessionReachable(req.TTYWatchHome, sessionID),
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else if !timedOut {
			return nil, runErr
		}
	}
	return resp, nil
}

func waitWatchDetachAfterGrokModes(t *testing.T, req *Request, sessionID string, modesWait time.Duration) (*Response, error) {
	t.Helper()
	return waitWatchDetachAfterGrokModesWithKey(t, req, sessionID, modesWait, []byte(kittyCtrlC))
}

func readPTYUntilGrokModes(ptmx *os.File, timeout time.Duration) (string, bool) {
	var buf bytes.Buffer
	deadline := time.Now().Add(timeout)
	tmp := make([]byte, 4096)
	for time.Now().Before(deadline) {
		_ = ptmx.SetReadDeadline(time.Now().Add(150 * time.Millisecond))
		n, err := ptmx.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])
			s := buf.String()
			if strings.Contains(s, "\x1b[?1049h") && strings.Contains(s, "\x1b[?u") {
				return s, true
			}
		}
		if err != nil {
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				continue
			}
			if err == io.EOF {
				break
			}
			break
		}
	}
	return buf.String(), false
}

func phaseWatchCtrlCDetachesNonRawStdin(t *testing.T, req *Request) (*Response, error) {
	sessionID := StartDetachedSession(t, req)

	ptm, tty, err := pty.Open()
	if err != nil {
		return nil, err
	}
	defer ptm.Close()
	defer tty.Close()

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(req.Bin, "watch", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)
	cmd.Stdout = tty
	cmd.Stdin = stdinR
	if err := cmd.Start(); err != nil {
		stdinR.Close()
		stdinW.Close()
		return nil, err
	}
	_ = stdinR.Close()
	defer stdinW.Close()

	return waitWatchDetachAfterInput(t, req, sessionID, func(cmd *exec.Cmd, _ *os.File) error {
		return cmd.Process.Signal(syscall.SIGINT)
	}, cmd)
}

func waitWatchDetachAfterInput(
	t *testing.T,
	req *Request,
	sessionID string,
	deliver func(cmd *exec.Cmd, ptmx *os.File) error,
	startedCmd *exec.Cmd,
) (*Response, error) {
	t.Helper()

	var cmd *exec.Cmd
	var ptmx *os.File
	if startedCmd != nil {
		cmd = startedCmd
	} else {
		var err error
		cmd = exec.Command(req.Bin, "watch", sessionID)
		cmd.Env = envWithHome(req.TTYWatchHome)
		ptmx, err = pty.Start(cmd)
		if err != nil {
			return nil, err
		}
	}

	time.Sleep(500 * time.Millisecond)
	if err := deliver(cmd, ptmx); err != nil {
		terminateProcess(cmd)
		if ptmx != nil {
			_ = ptmx.Close()
		}
		return nil, err
	}

	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()

	var runErr error
	timedOut := false
	select {
	case runErr = <-waitDone:
	case <-time.After(3 * time.Second):
		timedOut = true
		terminateProcess(cmd)
		runErr = <-waitDone
	}
	if ptmx != nil {
		_ = ptmx.Close()
	}

	resp := &Response{
		SessionID:      sessionID,
		TimedOut:       timedOut,
		RegistryExists: RegistryExists(req.TTYWatchHome, sessionID),
		SessionRunning: SessionReachable(req.TTYWatchHome, sessionID),
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else if !timedOut {
			return nil, runErr
		}
	}
	return resp, nil
}

func phaseWatchReadonly(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"cat"}
	sessionID := StartDetachedSession(t, &detachReq)

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, "watch", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	cmd.Stdin = stdinR
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Start(); err != nil {
		stdinR.Close()
		stdinW.Close()
		return nil, err
	}
	time.Sleep(500 * time.Millisecond)
	_, _ = stdinW.Write([]byte("SHOULD_NOT_ECHO\n"))
	_ = stdinW.Close()
	_ = stdinR.Close()

	_ = cmd.Wait()
	combined := out.String()
	return &Response{
		Combined: combined,
		Stdout:   combined,
		SessionID: sessionID,
	}, nil
}

func phaseSnapshotSanitize(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", `printf '\033[31mRED\033[0m\nPLAIN_LINE\n'; sleep 300`}
	sessionID := StartDetachedSession(t, &detachReq)
	time.Sleep(500 * time.Millisecond)

	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"snapshot", sessionID})
	if err != nil {
		return nil, err
	}
	text := stdout
	if text == "" {
		text = stderr
	}
	return &Response{
		SessionID:      sessionID,
		SnapshotText:   text,
		ContainsEscape: ContainsANSIEscape(text),
		Stdout:         stdout,
		Stderr:         stderr,
		ExitCode:       code,
	}, nil
}

func phaseSnapshotMissing(t *testing.T, req *Request) (*Response, error) {
	id := req.SnapshotID
	if id == "" {
		id = "session-99999"
	}
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"snapshot", id})
	if err != nil {
		return nil, err
	}
	return &Response{
		Stdout:   stdout,
		Stderr:   stderr,
		Combined: combineOutput(stdout, stderr),
		ExitCode: code,
	}, nil
}

func phaseKillStop(t *testing.T, req *Request) (*Response, error) {
	sessionID := StartDetachedSession(t, req)
	if !SessionReachable(req.TTYWatchHome, sessionID) {
		return nil, fmt.Errorf("session %s not reachable before kill", sessionID)
	}
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"kill", sessionID})
	if err != nil {
		return nil, err
	}
	time.Sleep(300 * time.Millisecond)
	return &Response{
		SessionID:      sessionID,
		Stdout:         stdout,
		Stderr:         stderr,
		Combined:       combineOutput(stdout, stderr),
		ExitCode:       code,
		RegistryExists: RegistryExists(req.TTYWatchHome, sessionID),
		SessionRunning: SessionReachable(req.TTYWatchHome, sessionID),
	}, nil
}

func phaseKillMissing(t *testing.T, req *Request) (*Response, error) {
	id := req.KillID
	if id == "" {
		id = "session-99999"
	}
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"kill", id})
	if err != nil {
		return nil, err
	}
	return &Response{
		Stdout:   stdout,
		Stderr:   stderr,
		Combined: combineOutput(stdout, stderr),
		ExitCode: code,
	}, nil
}

func phaseKillStale(t *testing.T, req *Request) (*Response, error) {
	const staleID = "session-stale-1"
	if err := WriteStaleRegistry(req.TTYWatchHome, staleID, "127.0.0.1:1"); err != nil {
		return nil, err
	}
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"kill", staleID})
	if err != nil {
		return nil, err
	}
	return &Response{
		SessionID:      staleID,
		Stdout:         stdout,
		Stderr:         stderr,
		Combined:       combineOutput(stdout, stderr),
		ExitCode:       code,
		RegistryExists: RegistryExists(req.TTYWatchHome, staleID),
	}, nil
}

func phaseErrorCmd(t *testing.T, req *Request) (*Response, error) {
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"not-a-real-subcommand"})
	if err != nil {
		return nil, err
	}
	return &Response{
		Stdout:   stdout,
		Stderr:   stderr,
		Combined: combineOutput(stdout, stderr),
		ExitCode: code,
	}, nil
}

type ptyOpts struct {
	detachAfter time.Duration
	signalAfter time.Duration
	signalByte  byte
	writeAfter  time.Duration
	writeBytes  []byte
	readUntil   string
	maxWait     time.Duration
}

func execPipeStdinSession(t *testing.T, req *Request, argv []string, opts ptyOpts) (*Response, error) {
	cmd := exec.Command(req.Bin, withRunSubcommand(argv)...)
	cmd.Env = envWithHome(req.TTYWatchHome)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	start := time.Now()
	readBudget := 12 * time.Second
	if opts.maxWait > 0 {
		readBudget = opts.maxWait + 2*time.Second
	}

	var output bytes.Buffer
	readDone := make(chan struct{})
	readMatched := make(chan struct{})
	go func() {
		defer close(readDone)
		buf := make([]byte, 4096)
		deadline := start.Add(readBudget)
		for time.Now().Before(deadline) {
			n, readErr := stdout.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
				if opts.readUntil != "" && strings.Contains(output.String(), opts.readUntil) {
					close(readMatched)
					break
				}
			}
			if readErr != nil {
				break
			}
		}
	}()

	if opts.writeAfter > 0 {
		time.Sleep(opts.writeAfter)
		if len(opts.writeBytes) > 0 {
			_, _ = stdin.Write(opts.writeBytes)
		}
	}

	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()

	var runErr error
	timedOut := false
	waitLimit := readBudget
	if opts.maxWait > 0 {
		waitLimit = opts.maxWait
	}
	select {
	case <-readMatched:
		terminateProcess(cmd)
		runErr = <-waitDone
	case runErr = <-waitDone:
	case <-time.After(waitLimit):
		timedOut = true
		terminateProcess(cmd)
		runErr = <-waitDone
	}
	_ = stdin.Close()
	<-readDone

	resp := &Response{
		Stdout:   output.String(),
		Combined: strings.TrimRight(output.String(), "\n"),
		TimedOut: timedOut,
		Elapsed:  time.Since(start),
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else if !timedOut {
			return nil, runErr
		}
	}
	return resp, nil
}

func execPTYSession(t *testing.T, req *Request, argv []string, opts ptyOpts) (*Response, error) {
	cmd := exec.Command(req.Bin, withRunSubcommand(argv)...)
	cmd.Env = envWithHome(req.TTYWatchHome)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	readBudget := 12 * time.Second
	if opts.maxWait > 0 {
		readBudget = opts.maxWait + 2*time.Second
	}

	var output bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		buf := make([]byte, 4096)
		deadline := start.Add(readBudget)
		for time.Now().Before(deadline) {
			n, readErr := ptmx.Read(buf)
			if n > 0 {
				output.Write(buf[:n])
				if opts.readUntil != "" && strings.Contains(output.String(), opts.readUntil) {
					break
				}
			}
			if readErr != nil {
				break
			}
			if opts.readUntil == "" && opts.detachAfter == 0 && opts.signalAfter == 0 {
				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	if opts.signalAfter > 0 {
		time.Sleep(opts.signalAfter)
		if opts.signalByte != 0 {
			_, _ = ptmx.Write([]byte{opts.signalByte})
		}
	}
	if opts.writeAfter > 0 {
		time.Sleep(opts.writeAfter)
		if len(opts.writeBytes) > 0 {
			_, _ = ptmx.Write(opts.writeBytes)
		}
	}
	if opts.detachAfter > 0 {
		time.Sleep(opts.detachAfter)
		_, _ = ptmx.Write([]byte{0x1d})
	}

	waitDone := make(chan error, 1)
	go func() { waitDone <- cmd.Wait() }()

	var runErr error
	timedOut := false
	waitLimit := readBudget
	if opts.maxWait > 0 {
		waitLimit = opts.maxWait
	}
	select {
	case runErr = <-waitDone:
	case <-time.After(waitLimit):
		timedOut = true
		terminateProcess(cmd)
		runErr = <-waitDone
	}
	_ = ptmx.Close()
	<-readDone

	resp := &Response{
		Stdout:   output.String(),
		Combined: strings.TrimRight(output.String(), "\n"),
		TimedOut: timedOut,
		Elapsed:  time.Since(start),
	}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			resp.ExitCode = exitErr.ExitCode()
		} else if !timedOut {
			return nil, runErr
		}
	}
	return resp, nil
}

func runCLI(bin, home string, args []string) (string, int, error) {
	stdout, stderr, code, err := runCLISeparate(bin, home, args)
	if err != nil {
		return "", code, err
	}
	return combineOutput(stdout, stderr), code, nil
}

func runCLISeparate(bin, home string, args []string) (stdout, stderr string, exitCode int, err error) {
	cmd := exec.Command(bin, args...)
	cmd.Env = envWithHome(home)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return "", "", 0, runErr
		}
	}
	return outBuf.String(), errBuf.String(), exitCode, nil
}

func envWithHome(home string) []string {
	env := os.Environ()
	prefix := "TTY_WATCH_HOME="
	out := make([]string, 0, len(env)+1)
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			continue
		}
		out = append(out, e)
	}
	out = append(out, prefix+home)
	return out
}

func waitForRegistrySession(home string, timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ids, err := ListRegistryIDs(home)
		if err != nil {
			return "", err
		}
		if len(ids) > 0 {
			return ids[0], nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return "", fmt.Errorf("timeout waiting for registry session under %s", RegistryDir(home))
}

func waitPTYClientExit(cmd *exec.Cmd, timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		terminateProcess(cmd)
	}
}

func terminateProcess(cmd *exec.Cmd) {
	if cmd == nil || cmd.Process == nil {
		return
	}
	_ = cmd.Process.Signal(syscall.SIGTERM)
	time.Sleep(100 * time.Millisecond)
	_ = cmd.Process.Kill()
}

func terminateProcessByHome(home, bin string) {
	ids, _ := ListRegistryIDs(home)
	for _, id := range ids {
		_, _, _, _ = runCLISeparate(bin, home, []string{"kill", id})
	}
}

func combineOutput(stdout, stderr string) string {
	stdout = strings.TrimSpace(stdout)
	stderr = strings.TrimSpace(stderr)
	switch {
	case stdout != "" && stderr != "":
		return stdout + "\n" + stderr
	case stderr != "":
		return stderr
	default:
		return stdout
	}
}

func tcpReachable(addr string) bool {
	conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

func findModuleRoot() (string, error) {
	start := os.Getenv("DOCTEST_ROOT")
	if start == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		start = wd
	}
	for dir := start; ; dir = filepath.Dir(dir) {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "script/tty-watch")); err == nil {
				return dir, nil
			}
		}
		if filepath.Dir(dir) == dir {
			return "", fmt.Errorf("could not find module root with script/tty-watch above %s", start)
		}
	}
}

// DrainPTY reads from ptmx until idle or timeout (test helper).
func DrainPTY(ptmx *os.File, timeout time.Duration) string {
	var buf bytes.Buffer
	deadline := time.Now().Add(timeout)
	tmp := make([]byte, 4096)
	for time.Now().Before(deadline) {
		remain := time.Until(deadline)
		if remain <= 0 {
			break
		}
		if remain > 50*time.Millisecond {
			remain = 50 * time.Millisecond
		}
		_ = ptmx.SetReadDeadline(time.Now().Add(remain))
		n, err := ptmx.Read(tmp)
		_ = ptmx.SetReadDeadline(time.Time{})
		if n > 0 {
			buf.Write(tmp[:n])
		}
		if err != nil {
			if ne, ok := err.(interface{ Timeout() bool }); ok && ne.Timeout() {
				continue
			}
			if err == io.EOF {
				break
			}
			break
		}
	}
	return buf.String()
}
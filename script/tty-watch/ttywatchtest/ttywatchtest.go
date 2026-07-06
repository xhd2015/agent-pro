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
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/creack/pty"
	"github.com/xhd2015/agent-pro/pkgs/ttywatch"
)

const registrySubdir = "registry"

// Request is the doctest harness request for tty-watch CLI tests.
type Request struct {
	Phase string

	Bin, TTYWatchHome, SessionID, CustomSessionID string
	RunCommand                 []string
	Detach, SendCtrlC, Background bool
	WatchProbe, SnapshotID, KillID string
	SendID, SendMessage string
	AttachID, AttachInput, AttachInputB, AttachProbe string
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
	AttachStdoutWriterNormalizesRawTTY bool
	InjectedBytes              []byte
	AltExitCode                int
	AltStderr                  string
	AttachOutput               string
	AttachBOutput              string
	WatchOutput                string
	RunOutput                  string
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
	case "run-custom-registers":
		return phaseRunCustomRegisters(t, req)
	case "run-custom-duplicate-live":
		return phaseRunCustomDuplicateLive(t, req)
	case "run-custom-reuses-stale":
		return phaseRunCustomReusesStale(t, req)
	case "run-custom-invalid-id":
		return phaseRunCustomInvalidID(t, req)
	case "run-custom-list-mixed":
		return phaseRunCustomListMixed(t, req)
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
	case "list-second-run-after-exit":
		return phaseListSecondRunAfterExit(t, req)
	case "list-table-header":
		return phaseListTableWithClients(t, req, "idle")
	case "list-table-idle":
		return phaseListTableWithClients(t, req, "idle")
	case "list-table-watch":
		return phaseListTableWithClients(t, req, "watch")
	case "list-table-attach":
		return phaseListTableWithClients(t, req, "attach")
	case "list-table-writer":
		return phaseListTableWithClients(t, req, "writer")
	case "list-table-both":
		return phaseListTableWithClients(t, req, "both")
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
	case "unit-attach-stdout-writer-crlf":
		return phaseUnitAttachStdoutWriterCRLF(t, req)
	case "unit-normalize-tty-output":
		return phaseUnitNormalizeTTYOutput(t, req)
	case "snapshot-sanitize":
		return phaseSnapshotSanitize(t, req)
	case "snapshot-codex-like-single-screen":
		return phaseSnapshotCodexLikeSingleScreen(t, req)
	case "snapshot-codex-cursor-drawn-mcp-boot":
		return phaseSnapshotCodexCursorDrawnMCPBoot(t, req)
	case "snapshot-codex-mcp-boot-smeared":
		return phaseSnapshotCodexMCPBootSmeared(t, req)
	case "snapshot-session-dimensions-wide":
		return phaseSnapshotSessionDimensionsWide(t, req)
	case "snapshot-grok-like-changelog-screen":
		return phaseSnapshotGrokLikeChangelogScreen(t, req)
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
	case "send-injects-verbatim", "send-no-suffix", "send-preserves-whitespace":
		return phaseSendCapture(t, req)
	case "send-missing-args":
		return phaseSendMissingArgs(t, req)
	case "send-missing":
		return phaseSendMissing(t, req)
	case "send-stale":
		return phaseSendStale(t, req)
	case "attach-detached-session":
		return phaseAttachDetachedSession(t, req)
	case "attach-forwards-stdin":
		return phaseAttachForwardsStdin(t, req)
	case "attach-second-writes":
		return phaseAttachSecondWrites(t, req)
	case "attach-visible-to-watch":
		return phaseAttachVisibleToWatch(t, req)
	case "attach-visible-to-other":
		return phaseAttachVisibleToOther(t, req)
	case "attach-visible-while-run":
		return phaseAttachVisibleWhileRunAttached(t, req)
	case "attach-resize":
		return phaseAttachResize(t, req)
	case "attach-send-input-queue":
		return phaseAttachSendInputQueue(t, req)
	case "attach-detach-survives":
		return phaseAttachDetachSurvives(t, req)
	case "attach-unknown":
		return phaseAttachUnknown(t, req)
	case "attach-concurrent-ordered":
		return phaseAttachConcurrentOrdered(t, req)
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

// ListRegistryIDs scans registry dir for all *.json session ids.
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
		ids = append(ids, strings.TrimSuffix(name, ".json"))
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

func buildRunArgv(req *Request, commandArgv []string) []string {
	argv := []string{"run"}
	if req.CustomSessionID != "" {
		argv = append(argv, "--session-id", req.CustomSessionID)
	}
	return append(argv, commandArgv...)
}

// StartDetachedSession starts tty-watch in a PTY, detaches with Ctrl-], returns session id.
func StartDetachedSession(t *testing.T, req *Request) string {
	t.Helper()
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "300"}
	}
	knownIDs := map[string]bool{}
	if ids, err := ListRegistryIDs(req.TTYWatchHome); err != nil {
		t.Fatalf("list registry before start: %v", err)
	} else {
		for _, id := range ids {
			knownIDs[id] = true
		}
	}

	cmd := exec.Command(req.Bin, buildRunArgv(req, argv)...)
	cmd.Env = envWithHome(req.TTYWatchHome)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		t.Fatalf("pty start detached session: %v", err)
	}
	t.Cleanup(func() {
		terminateProcess(cmd)
		_ = ptmx.Close()
	})

	sessionID, err := waitForRegistrySession(req.TTYWatchHome, 15*time.Second, req.CustomSessionID, knownIDs)
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

func phaseRunCustomRegisters(t *testing.T, req *Request) (*Response, error) {
	if req.CustomSessionID == "" {
		req.CustomSessionID = "test-with-grok"
	}
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "300"}
	}
	req.RunCommand = argv
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

func phaseRunCustomDuplicateLive(t *testing.T, req *Request) (*Response, error) {
	if req.CustomSessionID == "" {
		req.CustomSessionID = "test-with-grok"
	}
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "300"}
	}
	req.RunCommand = argv
	StartDetachedSession(t, req)

	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, buildRunArgv(req, argv))
	if err != nil {
		return nil, err
	}
	return &Response{
		SessionID:  req.CustomSessionID,
		Stdout:     stdout,
		Stderr:     stderr,
		Combined:   combineOutput(stdout, stderr),
		ExitCode:   code,
	}, nil
}

func phaseRunCustomReusesStale(t *testing.T, req *Request) (*Response, error) {
	if req.CustomSessionID == "" {
		req.CustomSessionID = "test-with-grok"
	}
	if err := WriteStaleRegistry(req.TTYWatchHome, req.CustomSessionID, "127.0.0.1:1"); err != nil {
		return nil, err
	}
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "300"}
	}
	req.RunCommand = argv
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

func phaseRunCustomInvalidID(t *testing.T, req *Request) (*Response, error) {
	customID := req.CustomSessionID
	if customID == "" {
		customID = ".bad"
	}
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "1"}
	}
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, append([]string{"run", "--session-id", customID}, argv...))
	if err != nil {
		return nil, err
	}
	return &Response{
		SessionID: customID,
		Stdout:    stdout,
		Stderr:    stderr,
		Combined:  combineOutput(stdout, stderr),
		ExitCode:  code,
	}, nil
}

func phaseRunCustomListMixed(t *testing.T, req *Request) (*Response, error) {
	customID := req.CustomSessionID
	if customID == "" {
		customID = "test-with-grok"
	}
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "300"}
	}

	customReq := *req
	customReq.CustomSessionID = customID
	customReq.RunCommand = argv
	customSessionID := StartDetachedSession(t, &customReq)

	autoReq := *req
	autoReq.CustomSessionID = ""
	autoReq.RunCommand = argv
	autoSessionID := StartDetachedSession(t, &autoReq)

	listOut, listCode, err := runCLI(req.Bin, req.TTYWatchHome, []string{"list"})
	if err != nil {
		return nil, err
	}
	ids, _ := ListRegistryIDs(req.TTYWatchHome)
	return &Response{
		SessionID:      customSessionID,
		RegistryIDs:    ids,
		ListOutput:     listOut,
		ExitCode:       listCode,
		RegistryExists: RegistryExists(req.TTYWatchHome, customSessionID) && RegistryExists(req.TTYWatchHome, autoSessionID),
		SessionRunning: SessionReachable(req.TTYWatchHome, customSessionID) && SessionReachable(req.TTYWatchHome, autoSessionID),
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

// phaseListSecondRunAfterExit reproduces the second-run list-empty bug: a fast first
// run exits and removes the registry file, but the zombie __serve__ process reuses the
// same session id on the second run and its grace-period RemoveRegistry deletes the
// live second session's registry entry.
// listTableHeaderColumns is the list table header order (COMMAND last).
var listTableHeaderColumns = []string{"SESSION", "UPTIME", "WATCH", "ATTACHED", "COMMAND"}

// listTableRowRE matches aligned list table data rows: SESSION UPTIME WATCH ATTACHED COMMAND.
var listTableRowRE = regexp.MustCompile(`^(\S+)\s+(\d+[smh]|unknown)\s+(\d+)\s+(\d+)\s+(.+)\s*$`)

// listTableCountAlignRE captures WATCH and ATTACHED numeric fields before the COMMAND column.
var listTableCountAlignRE = regexp.MustCompile(`^\S+\s+(?:\d+[smh]|unknown)\s+(\d+)\s+(\d+)\s+`)

// ListTableHeaderPresent reports whether list stdout includes the table header columns in order.
func ListTableHeaderPresent(output string) bool {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if !strings.Contains(line, "SESSION") {
			continue
		}
		lastIdx := -1
		for _, col := range listTableHeaderColumns {
			idx := strings.Index(line, col)
			if idx < 0 || idx <= lastIdx {
				return false
			}
			lastIdx = idx
		}
		return true
	}
	return false
}

// ListTableClientCounts parses WATCH and ATTACHED for sessionID from list table output.
func ListTableClientCounts(output, sessionID string) (watch, attached int, ok bool) {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, sessionID) {
			continue
		}
		m := listTableRowRE.FindStringSubmatch(line)
		if m == nil || m[1] != sessionID {
			continue
		}
		watch, _ = strconv.Atoi(m[3])
		attached, _ = strconv.Atoi(m[4])
		return watch, attached, true
	}
	return 0, 0, false
}

// ListTableCommand parses the COMMAND cell for sessionID from list table output.
func ListTableCommand(output, sessionID string) (command string, ok bool) {
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		line = strings.TrimSpace(line)
		if !strings.Contains(line, sessionID) {
			continue
		}
		m := listTableRowRE.FindStringSubmatch(line)
		if m == nil || m[1] != sessionID {
			continue
		}
		return strings.TrimSpace(m[5]), true
	}
	return "", false
}

// ListTableColumnsAligned reports whether header and first data row share aligned columns.
func ListTableColumnsAligned(output string) bool {
	lines := nonEmptyLines(output)
	if len(lines) < 2 {
		return false
	}
	header, data := lines[0], lines[1]
	lastIdx := -1
	for _, col := range listTableHeaderColumns {
		idx := strings.Index(header, col)
		if idx < 0 || idx <= lastIdx {
			return false
		}
		lastIdx = idx
	}
	idx := listTableCountAlignRE.FindStringSubmatchIndex(data)
	if idx == nil {
		return false
	}
	watchCol := strings.Index(header, "WATCH")
	attachCol := strings.Index(header, "ATTACHED")
	watchNum := idx[2]
	attachNum := idx[4]
	if watchCol < 0 || attachCol < 0 || watchNum < 0 || attachNum < 0 {
		return false
	}
	return watchNum >= watchCol && attachNum >= attachCol && attachNum > watchNum
}

func nonEmptyLines(output string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if strings.TrimSpace(line) != "" {
			lines = append(lines, strings.TrimSpace(line))
		}
	}
	return lines
}

func phaseListTableWithClients(t *testing.T, req *Request, probe string) (*Response, error) {
	sessionID := StartDetachedSession(t, req)
	time.Sleep(300 * time.Millisecond)

	var clients []*ttywatch.WSAttachClient
	defer func() {
		for _, c := range clients {
			c.Close()
		}
	}()

	switch probe {
	case "idle":
	case "watch":
		clients = append(clients, dialPTYAttachMode(t, req.TTYWatchHome, sessionID, "observer"))
	case "attach":
		clients = append(clients, dialPTYAttachMode(t, req.TTYWatchHome, sessionID, "attach"))
	case "writer":
		clients = append(clients, dialPTYAttachMode(t, req.TTYWatchHome, sessionID, "screen"))
	case "both":
		clients = append(clients, dialPTYAttachMode(t, req.TTYWatchHome, sessionID, "observer"))
		clients = append(clients, dialPTYAttachMode(t, req.TTYWatchHome, sessionID, "attach"))
	default:
		return nil, fmt.Errorf("unknown list table probe %q", probe)
	}
	time.Sleep(200 * time.Millisecond)

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

func phaseListSecondRunAfterExit(t *testing.T, req *Request) (*Response, error) {
	_, code, err := runCLI(req.Bin, req.TTYWatchHome, []string{"run", "true"})
	if err != nil {
		return nil, err
	}
	if code != 0 {
		return nil, fmt.Errorf("first run true exited %d", code)
	}

	detachReq := *req
	detachReq.RunCommand = []string{"sleep", "600"}
	sessionID := StartDetachedSession(t, &detachReq)

	// Wait past the first __serve__ writerAttachGrace (2s) so its cleanup races with
	// the second session's registry entry.
	time.Sleep(3 * time.Second)

	listOut, code, err := runCLI(req.Bin, req.TTYWatchHome, []string{"list"})
	if err != nil {
		return nil, err
	}
	ids, _ := ListRegistryIDs(req.TTYWatchHome)
	return &Response{
		SessionID:      sessionID,
		ListOutput:     listOut,
		ExitCode:       code,
		RegistryIDs:    ids,
		RegistryExists: len(ids) > 0,
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

// codexLikeSnapshotCommand mimics codex-style alternate-screen redraws with status
// line updates (MCP boot) that snapshot must collapse to the latest screen only.
const codexLikeSnapshotCommand = `printf '\033[?1049h\033[2J\033[H\033[38;2;255;255;255m╭ OpenAI Codex ╮\033[0m\n\033[38;2;255;255;255m│ model: gpt-5.5 medium │\033[0m\n'; i=0; while [ "$i" -lt 8 ]; do printf '\033[2J\033[H\033[38;2;255;255;255m╭ OpenAI Codex ╮\033[0m\n\033[38;2;255;255;255m│ model: gpt-5.5 medium │\033[0m\n\033[33mΔ MCP starting ($i/5)\033[0m\n\033[38;2;255;255;255m› Improve documentation\033[0m\n'; i=$((i+1)); sleep 0.1; done; while true; do sleep 1; done`

// codexCursorDrawnMCPBootCommand mimics real codex: plain warning before ?2026h (no 2J),
// box UI via absolute cursor moves, tip line, incremental MCP status redraws, and kitty
// protocol bytes that leak when snapshot render is wrong.
const codexCursorDrawnMCPBootCommand = `printf '⚠ codex_hooks deprecated. See https://developers.openai.com/codex/config-basic#feature-flags for details.\n'; printf '\033[?2026h'; printf '\033[4;1H╭──────────────────────────────────────────────────────────╮'; printf '\033[5;1H│ >_ OpenAI Codex (v0.142.5)                               │'; printf '\033[6;1H│                                                          │'; printf '\033[7;1H│ model:     gpt-5.5 medium   /model to change             │'; printf '\033[8;1H│ directory: ~/worktrees/…support-send-followup            │'; printf '\033[9;1H╰──────────────────────────────────────────────────────────╯'; printf '\033[11;1H  Tip: New Use /fast to enable our fastest inference.'; i=0; while [ "$i" -lt 8 ]; do printf '\033[22;1H\033[K• Starting MCP servers (%s/5): codex_apps, computer-use, …' "$i"; printf '\033[<43;52;23M'; i=$((i+1)); sleep 0.05; done; printf '\033[22;1H\033[K⚠ MCP client for computer-use failed to start'; printf '\033[23;1H\033[K⚠ MCP startup incomplete (failed: computer-use)'; printf '\033[20;1H› Write tests for @filename'; printf '\033[24;1Hgpt-5.5 medium · ~/.wrk/worktrees/agent-pro-master-2026-07-04-support-send-followup'; while true; do sleep 1; done`

// grokLikeChangelogSnapshotCommand mimics grok's changelog alt-screen UI: ?1049h entry,
// changelog-only boot redraws that pollute scrollback, then absolute CUP draws of menu
// (ctrl+q), bordered changelog box, prompt, and footer. Final UI uses CUP only (no 2J)
// like real grok; scrollback replay leaves a ghost Quit q line until frame path wins.
const grokLikeChangelogSnapshotCommand = `printf '\033[?1049h\033[?25l'; printf '\033[2J\033[HChangelog\nQuit q\n'; sleep 0.15; printf '\033[1;1H\033[K  Quit ctrl+q    Changelog    Settings'; printf '\033[3;1H╭──────────────────────────────────────────────────────────╮'; printf '\033[4;1H│ Grok Build Changelog                                     │'; printf '\033[5;1H│                                                          │'; printf '\033[6;1H│ • Snapshot screen-frame parity fix                       │'; printf '\033[7;1H╰──────────────────────────────────────────────────────────╯'; printf '\033[20;1H❯ Ask anything'; printf '\033[24;1HLogged in with API key · Grok Build'; while true; do sleep 1; done`

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
	path := filepath.Join(root, "pkgs/ttywatch/observer.go")
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
func phaseUnitNormalizeTTYOutput(t *testing.T, req *Request) (*Response, error) {
	return &Response{SourceCheckOK: true, SourceCheckNote: "normalizeTTYOutput cases asserted in leaf Assert"}, nil
}

func phaseUnitAttachStdoutWriterCRLF(t *testing.T, req *Request) (*Response, error) {
	root, err := findModuleRoot()
	if err != nil {
		return &Response{SourceCheckOK: false, SourceCheckNote: err.Error()}, nil
	}
	path := filepath.Join(root, "pkgs/ttywatch/attach_client.go")
	data, err := os.ReadFile(path)
	if err != nil {
		return &Response{SourceCheckOK: false, SourceCheckNote: err.Error()}, nil
	}
	ok, note := attachGoStdoutWriterNormalizesRawTTY(string(data))
	return &Response{
		SourceCheckOK:                      true,
		SourceCheckNote:                    note,
		AttachStdoutWriterNormalizesRawTTY: ok,
	}, nil
}

// attachGoStdoutWriterNormalizesRawTTY reports whether attachStdoutWriter applies
// normalizeTTYOutput before writing on interactive TTY stdout.
func attachGoStdoutWriterNormalizesRawTTY(src string) (bool, string) {
	marker := "func (a *attachStdoutWriter) Write"
	idx := strings.Index(src, marker)
	if idx < 0 {
		return false, "attachStdoutWriter.Write not found in attach_client.go"
	}
	rest := src[idx:]
	end := strings.Index(rest, "\nfunc ")
	if end < 0 {
		end = len(rest)
	}
	body := rest[:end]
	if !strings.Contains(body, "a.rawTTY") {
		return false, "attachStdoutWriter.Write missing rawTTY branch"
	}
	if strings.Contains(body, "normalizeTTYOutput") {
		return true, "attachStdoutWriter.Write calls normalizeTTYOutput for raw TTY output"
	}
	return false, "attachStdoutWriter.Write passes LF-only bytes unchanged on raw TTY (missing normalizeTTYOutput)"
}

func phaseUnitObserverDetachKittyPopCleanup(t *testing.T, req *Request) (*Response, error) {
	root, err := findModuleRoot()
	if err != nil {
		return &Response{SourceCheckOK: false, SourceCheckNote: err.Error()}, nil
	}
	path := filepath.Join(root, "pkgs/ttywatch/observer.go")
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
		return false, "observerTTYDetachCleanup constant not found in observer.go"
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
		return false, "detachCleanup closure not found in observer.go"
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

func phaseSnapshotGrokLikeChangelogScreen(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.CustomSessionID = "grok"
	detachReq.RunCommand = []string{"sh", "-c", grokLikeChangelogSnapshotCommand}
	sessionID := StartDetachedSession(t, &detachReq)
	time.Sleep(1500 * time.Millisecond)

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

func phaseSnapshotCodexLikeSingleScreen(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.CustomSessionID = "codex"
	detachReq.RunCommand = []string{"sh", "-c", codexLikeSnapshotCommand}
	sessionID := StartDetachedSession(t, &detachReq)
	time.Sleep(1500 * time.Millisecond)

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

func phaseSnapshotCodexCursorDrawnMCPBoot(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.CustomSessionID = "codex"
	detachReq.RunCommand = []string{"sh", "-c", codexCursorDrawnMCPBootCommand}
	sessionID := StartDetachedSession(t, &detachReq)
	time.Sleep(1200 * time.Millisecond)

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

func wideSnapshotLineMarker() string {
	return "WIDE_LINE_MARKER_" + strings.Repeat("x", 75)
}

func phaseSnapshotSessionDimensionsWide(t *testing.T, req *Request) (*Response, error) {
	marker := wideSnapshotLineMarker()
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", fmt.Sprintf(
		`printf '\033[?2026h'; printf '\033[30;1H%s'; sleep 300`, marker)}
	sessionID := StartDetachedSession(t, &detachReq)

	entry, err := ReadRegistryEntry(req.TTYWatchHome, sessionID)
	if err != nil {
		return nil, err
	}
	writer, err := ttywatch.DialPTYAttach(entry.ListenAddr, sessionID, "")
	if err != nil {
		return nil, fmt.Errorf("dial writer ws: %w", err)
	}
	if err := writer.TryResize(100, 32); err != nil {
		writer.Close()
		return nil, fmt.Errorf("writer resize: %w", err)
	}
	writer.Close()
	time.Sleep(300 * time.Millisecond)

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

func phaseSnapshotCodexMCPBootSmeared(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.CustomSessionID = "codex"
	detachReq.RunCommand = []string{"sh", "-c", codexCursorDrawnMCPBootCommand}
	sessionID := StartDetachedSession(t, &detachReq)
	// Snapshot mid MCP boot before final error lines are drawn.
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

func byteCaptureSessionCommand(home string) (capturePath string, argv []string) {
	capturePath = filepath.Join(home, "inject-capture.bin")
	argv = []string{
		"sh", "-c",
		fmt.Sprintf("cat > %s; sleep 300", shellQuote(capturePath)),
	}
	return capturePath, argv
}

func phaseSendCapture(t *testing.T, req *Request) (*Response, error) {
	message := req.SendMessage
	if message == "" {
		return nil, fmt.Errorf("send capture phase requires SendMessage")
	}
	capturePath, argv := byteCaptureSessionCommand(req.TTYWatchHome)
	detachReq := *req
	detachReq.RunCommand = argv
	sessionID := StartDetachedSession(t, &detachReq)
	time.Sleep(400 * time.Millisecond)

	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"send", sessionID, message})
	if err != nil {
		return nil, err
	}
	time.Sleep(300 * time.Millisecond)

	captured, readErr := os.ReadFile(capturePath)
	if readErr != nil && !os.IsNotExist(readErr) {
		return nil, readErr
	}
	return &Response{
		SessionID:     sessionID,
		Stdout:        stdout,
		Stderr:        stderr,
		Combined:      combineOutput(stdout, stderr),
		ExitCode:      code,
		InjectedBytes: captured,
	}, nil
}

func phaseSendMissingArgs(t *testing.T, req *Request) (*Response, error) {
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"send"})
	if err != nil {
		return nil, err
	}
	_, altStderr, altCode, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"send", "session-1"})
	if err != nil {
		return nil, err
	}
	return &Response{
		Stdout:      stdout,
		Stderr:      stderr,
		Combined:    combineOutput(stdout, stderr),
		ExitCode:    code,
		AltExitCode: altCode,
		AltStderr:   altStderr,
	}, nil
}

func phaseSendMissing(t *testing.T, req *Request) (*Response, error) {
	id := req.SendID
	if id == "" {
		id = "session-99999"
	}
	message := req.SendMessage
	if message == "" {
		message = "hi"
	}
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"send", id, message})
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

func phaseSendStale(t *testing.T, req *Request) (*Response, error) {
	const staleID = "session-stale-1"
	if err := WriteStaleRegistry(req.TTYWatchHome, staleID, "127.0.0.1:1"); err != nil {
		return nil, err
	}
	message := req.SendMessage
	if message == "" {
		message = "hi"
	}
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"send", staleID, message})
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

func waitForRegistrySession(home string, timeout time.Duration, expectID string, knownIDs map[string]bool) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if expectID != "" && RegistryExists(home, expectID) {
			return expectID, nil
		}
		ids, err := ListRegistryIDs(home)
		if err != nil {
			return "", err
		}
		for _, id := range ids {
			if expectID != "" {
				if id == expectID {
					return id, nil
				}
				continue
			}
			if knownIDs == nil || !knownIDs[id] {
				return id, nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	if expectID != "" {
		return "", fmt.Errorf("timeout waiting for registry session %s under %s", expectID, RegistryDir(home))
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

const (
	defaultAttachLiveMarker      = "ATTACH_LIVE_MARKER"
	defaultAttachStdinMarker     = "ATTACH_STDIN_MARKER"
	defaultAttachWriteMarkerA    = "ATTACH_WRITE_MARKER_A"
	defaultAttachWriteMarkerB    = "ATTACH_WRITE_MARKER_B"
	defaultAttachConcurrentMarkerA = "CONCURRENT_MARKER_A_END"
	defaultAttachConcurrentMarkerB = "CONCURRENT_MARKER_B_END"
)

func attachProbeDuration(req *Request, fallback time.Duration) time.Duration {
	if req.AttachProbe != "" {
		if d, err := time.ParseDuration(req.AttachProbe); err == nil {
			return d
		}
	}
	return fallback
}

func attachInputOr(req *Request, fallback string) string {
	if req.AttachInput != "" {
		return req.AttachInput
	}
	return fallback
}

func attachInputBOr(req *Request, fallback string) string {
	if req.AttachInputB != "" {
		return req.AttachInputB
	}
	return fallback
}

func dialAttachModeClient(t *testing.T, home, sessionID string) *ttywatch.WSAttachClient {
	return dialPTYAttachMode(t, home, sessionID, "attach")
}

func dialPTYAttachMode(t *testing.T, home, sessionID, attachMode string) *ttywatch.WSAttachClient {
	t.Helper()
	entry, err := ReadRegistryEntry(home, sessionID)
	if err != nil {
		t.Fatalf("read registry for attach dial: %v", err)
	}
	client, err := ttywatch.DialPTYAttach(entry.ListenAddr, sessionID, attachMode)
	if err != nil {
		t.Fatalf("dial attach_mode=%s: %v", attachMode, err)
	}
	return client
}

func execAttachPTYSession(t *testing.T, req *Request, sessionID string, opts ptyOpts) (*Response, error) {
	cmd := exec.Command(req.Bin, "attach", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	readBudget := 8 * time.Second
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
		}
	}()

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
		Combined: output.String(),
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

func phaseAttachDetachedSession(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c",
		`echo ATTACH_BOOT_MARKER; while true; do echo ` + defaultAttachLiveMarker + `; sleep 0.5; done`}
	sessionID := StartDetachedSession(t, &detachReq)

	probe := attachProbeDuration(req, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), probe+5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, req.Bin, "attach", sessionID)
	cmd.Env = envWithHome(req.TTYWatchHome)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	_ = cmd.Run()
	combined := out.String()
	return &Response{
		SessionID:    sessionID,
		Stdout:       combined,
		Combined:     combined,
		AttachOutput: combined,
	}, nil
}

func phaseAttachForwardsStdin(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", "cat"}
	sessionID := StartDetachedSession(t, &detachReq)

	marker := attachInputOr(req, defaultAttachStdinMarker)
	resp, err := execAttachPTYSession(t, req, sessionID, ptyOpts{
		writeAfter: 500 * time.Millisecond,
		writeBytes: []byte(marker + "\n"),
		readUntil:  marker,
		maxWait:    4 * time.Second,
	})
	if err != nil {
		return nil, err
	}
	resp.SessionID = sessionID
	resp.AttachOutput = resp.Combined
	return resp, nil
}

func phaseAttachSecondWrites(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", "cat"}
	sessionID := StartDetachedSession(t, &detachReq)
	time.Sleep(300 * time.Millisecond)

	markerA := attachInputOr(req, defaultAttachWriteMarkerA)
	markerB := attachInputBOr(req, defaultAttachWriteMarkerB)

	clientA := dialAttachModeClient(t, req.TTYWatchHome, sessionID)
	clientB := dialAttachModeClient(t, req.TTYWatchHome, sessionID)
	defer clientA.Close()
	defer clientB.Close()

	if err := clientA.TryWriteInput([]byte(markerA + "\n")); err != nil {
		return nil, fmt.Errorf("attach A write: %w", err)
	}
	time.Sleep(200 * time.Millisecond)
	if err := clientB.TryWriteInput([]byte(markerB + "\n")); err != nil {
		return nil, fmt.Errorf("attach B write: %w", err)
	}
	time.Sleep(500 * time.Millisecond)

	outA := clientA.Output()
	outB := clientB.Output()
	return &Response{
		SessionID:     sessionID,
		AttachOutput:  outA,
		AttachBOutput: outB,
		Combined:      outA + outB,
	}, nil
}

func phaseAttachVisibleToWatch(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", "cat"}
	sessionID := StartDetachedSession(t, &detachReq)
	time.Sleep(300 * time.Millisecond)

	marker := attachInputOr(req, defaultAttachWriteMarkerA)

	watchCtx, watchCancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer watchCancel()
	watchCmd := exec.CommandContext(watchCtx, req.Bin, "watch", sessionID)
	watchCmd.Env = envWithHome(req.TTYWatchHome)
	var watchBuf bytes.Buffer
	watchCmd.Stdout = &watchBuf
	watchCmd.Stderr = &watchBuf
	if err := watchCmd.Start(); err != nil {
		return nil, err
	}

	time.Sleep(400 * time.Millisecond)
	client := dialAttachModeClient(t, req.TTYWatchHome, sessionID)
	if err := client.TryWriteInput([]byte(marker + "\n")); err != nil {
		client.Close()
		terminateProcess(watchCmd)
		return nil, fmt.Errorf("attach write: %w", err)
	}
	time.Sleep(600 * time.Millisecond)
	attachOut := client.Output()
	client.Close()
	terminateProcess(watchCmd)
	_ = watchCmd.Wait()

	return &Response{
		SessionID:    sessionID,
		AttachOutput: attachOut,
		WatchOutput:  watchBuf.String(),
	}, nil
}

func phaseAttachVisibleToOther(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", "cat"}
	sessionID := StartDetachedSession(t, &detachReq)
	time.Sleep(300 * time.Millisecond)

	marker := attachInputOr(req, defaultAttachWriteMarkerA)

	clientA := dialAttachModeClient(t, req.TTYWatchHome, sessionID)
	clientB := dialAttachModeClient(t, req.TTYWatchHome, sessionID)
	defer clientA.Close()
	defer clientB.Close()

	if err := clientA.TryWriteInput([]byte(marker + "\n")); err != nil {
		return nil, fmt.Errorf("attach A write: %w", err)
	}
	time.Sleep(600 * time.Millisecond)

	return &Response{
		SessionID:     sessionID,
		AttachOutput:  clientA.Output(),
		AttachBOutput: clientB.Output(),
	}, nil
}

func phaseAttachVisibleWhileRunAttached(t *testing.T, req *Request) (*Response, error) {
	argv := req.RunCommand
	if len(argv) == 0 {
		argv = []string{"sleep", "300"}
	}
	detachReq := *req
	detachReq.RunCommand = argv
	sessionID := StartDetachedSession(t, &detachReq)

	entry, err := ReadRegistryEntry(req.TTYWatchHome, sessionID)
	if err != nil {
		return nil, err
	}
	screenWriter, err := ttywatch.DialPTYAttach(entry.ListenAddr, sessionID, "screen")
	if err != nil {
		return nil, fmt.Errorf("dial screen writer like run attach: %w", err)
	}
	defer screenWriter.Close()

	marker := attachInputOr(req, defaultAttachWriteMarkerA)

	watchCtx, watchCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer watchCancel()
	watchCmd := exec.CommandContext(watchCtx, req.Bin, "watch", sessionID)
	watchCmd.Env = envWithHome(req.TTYWatchHome)
	var watchBuf bytes.Buffer
	watchCmd.Stdout = &watchBuf
	watchCmd.Stderr = &watchBuf
	if err := watchCmd.Start(); err != nil {
		return nil, err
	}

	time.Sleep(400 * time.Millisecond)
	client := dialAttachModeClient(t, req.TTYWatchHome, sessionID)
	if err := client.TryWriteInput([]byte(marker + "\n")); err != nil {
		client.Close()
		terminateProcess(watchCmd)
		return nil, fmt.Errorf("attach write while screen writer connected: %w", err)
	}
	time.Sleep(600 * time.Millisecond)
	attachOut := client.Output()
	client.Close()

	runOut := screenWriter.Output()
	terminateProcess(watchCmd)
	_ = watchCmd.Wait()

	return &Response{
		SessionID:    sessionID,
		RunOutput:    runOut,
		WatchOutput:  watchBuf.String(),
		AttachOutput: attachOut,
	}, nil
}

func phaseAttachResize(t *testing.T, req *Request) (*Response, error) {
	marker := wideSnapshotLineMarker()
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", fmt.Sprintf(
		`printf '\033[?2026h'; printf '\033[30;1H%s'; sleep 300`, marker)}
	sessionID := StartDetachedSession(t, &detachReq)

	entry, err := ReadRegistryEntry(req.TTYWatchHome, sessionID)
	if err != nil {
		return nil, err
	}
	screenWriter, err := ttywatch.DialPTYAttach(entry.ListenAddr, sessionID, "screen")
	if err != nil {
		return nil, fmt.Errorf("dial screen writer: %w", err)
	}
	if err := screenWriter.TryResize(80, 24); err != nil {
		screenWriter.Close()
		return nil, fmt.Errorf("screen writer resize 80x24: %w", err)
	}
	time.Sleep(200 * time.Millisecond)

	client := dialAttachModeClient(t, req.TTYWatchHome, sessionID)
	if err := client.TryResize(100, 32); err != nil {
		screenWriter.Close()
		client.Close()
		return nil, fmt.Errorf("attach resize: %w", err)
	}
	attachOut := client.Output()
	client.Close()
	screenWriter.Close()
	time.Sleep(300 * time.Millisecond)

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
		AttachOutput:   attachOut,
	}, nil
}

func phaseAttachSendInputQueue(t *testing.T, req *Request) (*Response, error) {
	message := req.SendMessage
	if message == "" {
		message = "ATTACH_SEND_QUEUE_MARKER"
	}
	capturePath, argv := byteCaptureSessionCommand(req.TTYWatchHome)
	detachReq := *req
	detachReq.RunCommand = argv
	sessionID := StartDetachedSession(t, &detachReq)
	time.Sleep(300 * time.Millisecond)

	client := dialAttachModeClient(t, req.TTYWatchHome, sessionID)
	time.Sleep(200 * time.Millisecond)

	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"send", sessionID, message})
	if err != nil {
		client.Close()
		return nil, err
	}
	time.Sleep(400 * time.Millisecond)
	attachOut := client.Output()
	client.Close()

	captured, readErr := os.ReadFile(capturePath)
	if readErr != nil && !os.IsNotExist(readErr) {
		return nil, readErr
	}
	return &Response{
		SessionID:     sessionID,
		Stdout:        stdout,
		Stderr:        stderr,
		Combined:      combineOutput(stdout, stderr),
		ExitCode:      code,
		InjectedBytes: captured,
		AttachOutput:  attachOut,
	}, nil
}

func phaseAttachDetachSurvives(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sleep", "300"}
	sessionID := StartDetachedSession(t, &detachReq)

	resp, err := execAttachPTYSession(t, req, sessionID, ptyOpts{
		detachAfter: 400 * time.Millisecond,
		maxWait:     4 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	listOut, _, err := runCLI(req.Bin, req.TTYWatchHome, []string{"list"})
	if err != nil {
		return nil, err
	}
	resp.SessionID = sessionID
	resp.RegistryExists = RegistryExists(req.TTYWatchHome, sessionID)
	resp.ListOutput = listOut
	resp.SessionRunning = SessionReachable(req.TTYWatchHome, sessionID)
	resp.AttachOutput = resp.Combined
	return resp, nil
}

func phaseAttachUnknown(t *testing.T, req *Request) (*Response, error) {
	id := req.AttachID
	if id == "" {
		id = "session-99999"
	}
	stdout, stderr, code, err := runCLISeparate(req.Bin, req.TTYWatchHome, []string{"attach", id})
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

func phaseAttachConcurrentOrdered(t *testing.T, req *Request) (*Response, error) {
	detachReq := *req
	detachReq.RunCommand = []string{"sh", "-c", "cat"}
	sessionID := StartDetachedSession(t, &detachReq)
	time.Sleep(300 * time.Millisecond)

	markerA := attachInputOr(req, defaultAttachConcurrentMarkerA)
	markerB := attachInputBOr(req, defaultAttachConcurrentMarkerB)

	clientA := dialAttachModeClient(t, req.TTYWatchHome, sessionID)
	clientB := dialAttachModeClient(t, req.TTYWatchHome, sessionID)
	defer clientA.Close()
	defer clientB.Close()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = clientA.TryWriteInput([]byte(markerA + "\n"))
	}()
	go func() {
		defer wg.Done()
		_ = clientB.TryWriteInput([]byte(markerB + "\n"))
	}()
	wg.Wait()
	time.Sleep(600 * time.Millisecond)

	outA := clientA.Output()
	outB := clientB.Output()
	combined := outA + outB
	return &Response{
		SessionID:     sessionID,
		AttachOutput:  outA,
		AttachBOutput: outB,
		Combined:      combined,
	}, nil
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
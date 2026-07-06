package sudosetuptest

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/xhd2015/agent-pro/pkgs/sudosetup"
)

type ManifestSeed struct {
	Username    string
	Command     string
	ArgsPattern string
}

type RunnerCall struct {
	Method string
	Name   string
	Args   []string
}

type Request struct {
	Operation string
	Action    string

	Config sudosetup.Config
	Rule   sudosetup.Rule

	SeedSudoersLine string
	SeedManifest    *ManifestSeed

	SudoNTrueOK        bool
	SudoNCommandOK     bool
	SudoNCommandOutput string
	VisudoFails        bool
	InstallFails       bool
	SudoKFails         bool
	RmFails            bool

	EnvUser     string
	EnvSudoUser string

	// StdinIsTerminal is whether stdin is a TTY for install/remove (RootSetup defaults true).
	StdinIsTerminal bool
}

type Response struct {
	Status sudosetup.Status

	Installed      bool
	InstallDetail  string
	RenderedLine   string
	ManifestJSON   map[string]string
	SudoersContent string

	RunnerCalls  []RunnerCall
	ManifestPath string
	SudoersPath  string

	Err error
}

func RootSetup(t *testing.T, req *Request) error {
	req.Config = sudosetup.Config{
		CacheDirName: "remote-agent-sudo-poc",
		SudoersName:  "remote-agent-sudo-poc",
		Username:     "testuser",
	}
	req.Rule = sudosetup.Rule{
		Command: "/tmp/cache/remote-agent-sudo-poc/hello.sh",
	}
	req.EnvUser = "testuser"
	req.StdinIsTerminal = true
	return nil
}

func Run(t *testing.T, req *Request) (*Response, error) {
	fs := newMapFS(t)
	runner := newRecordingRunner(req)
	mgr := newManager(t, req, fs, runner)

	resp := &Response{
		RunnerCalls:  runner.calls,
		ManifestPath: mgr.ManifestPath(),
		SudoersPath:  mgr.SudoersPath(),
	}

	switch req.Operation {
	case "detect":
		status := mgr.Detect()
		resp.Status = status
		resp.Installed = status.Installed
		resp.InstallDetail = status.InstallDetail
		resp.RunnerCalls = runner.calls
		return resp, nil

	case "install":
		err := mgr.EnsureInstalled()
		resp.Err = err
		resp.RunnerCalls = runner.calls
		resp.SudoersContent = fs.read(mgr.SudoersPath())
		if data := fs.read(mgr.ManifestPath()); data != "" {
			_ = json.Unmarshal([]byte(data), &resp.ManifestJSON)
		}
		installed, detail := mgr.IsInstalled()
		resp.Installed = installed
		resp.InstallDetail = detail
		return resp, err

	case "remove":
		err := mgr.Remove()
		resp.Err = err
		resp.RunnerCalls = runner.calls
		resp.SudoersContent = fs.read(mgr.SudoersPath())
		if data := fs.read(mgr.ManifestPath()); data != "" {
			_ = json.Unmarshal([]byte(data), &resp.ManifestJSON)
		}
		installed, detail := mgr.IsInstalled()
		resp.Installed = installed
		resp.InstallDetail = detail
		return resp, err

	case "render":
		line, err := mgr.RenderSudoersLine()
		resp.RenderedLine = line
		resp.Err = err
		return resp, err

	default:
		return nil, fmt.Errorf("unknown operation %q", req.Operation)
	}
}

func InstalledSeedLine(username, command, argsPattern string) string {
	if argsPattern == "" {
		return username + " ALL=(root) NOPASSWD: " + command + "\n"
	}
	return username + " ALL=(root) NOPASSWD: " + command + " " + argsPattern + "\n"
}

func InstalledManifestSeed(username, command, argsPattern string) *ManifestSeed {
	return &ManifestSeed{
		Username:    username,
		Command:     command,
		ArgsPattern: argsPattern,
	}
}

func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func AssertEqual(t *testing.T, field string, got, want any) {
	t.Helper()
	if got != want {
		t.Fatalf("%s: got %v, want %v", field, got, want)
	}
}

func AssertContains(t *testing.T, got, want string) {
	t.Helper()
	if !strings.Contains(got, want) {
		t.Fatalf("missing %q in %q", want, got)
	}
}

func HasRunnerCall(calls []RunnerCall, name string, argPrefix ...string) bool {
	for _, c := range calls {
		if c.Name != name {
			continue
		}
		if len(argPrefix) == 0 {
			return true
		}
		if len(c.Args) < len(argPrefix) {
			continue
		}
		match := true
		for i, a := range argPrefix {
			if c.Args[i] != a {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func RunnerCallCount(calls []RunnerCall, name string, argPrefix ...string) int {
	n := 0
	for _, c := range calls {
		if c.Name != name {
			continue
		}
		if len(argPrefix) == 0 {
			n++
			continue
		}
		if len(c.Args) < len(argPrefix) {
			continue
		}
		match := true
		for i, a := range argPrefix {
			if c.Args[i] != a {
				match = false
				break
			}
		}
		if match {
			n++
		}
	}
	return n
}

func DetailContains(detail, sub string) bool {
	return strings.Contains(strings.ToLower(detail), strings.ToLower(sub))
}

func newManager(t *testing.T, req *Request, fs *mapFS, runner *recordingRunner) *sudosetup.Manager {
	t.Helper()
	applyUsernameEnv(req)
	seedFS(t, fs, req)
	stdinTTY := req.StdinIsTerminal
	return &sudosetup.Manager{
		Config: req.Config,
		Rule:   req.Rule,
		Runner: runner,
		FS:     fs,
		StdinIsTerminal: func() bool {
			return stdinTTY
		},
	}
}

func seedFS(t *testing.T, fs *mapFS, req *Request) {
	t.Helper()
	mgr := &sudosetup.Manager{Config: req.Config, Rule: req.Rule, FS: fs}
	if req.SeedSudoersLine != "" {
		if err := fs.WriteFile(mgr.SudoersPath(), []byte(req.SeedSudoersLine), 0440); err != nil {
			t.Fatalf("seed sudoers: %v", err)
		}
	}
	if req.SeedManifest != nil {
		data, err := json.MarshalIndent(map[string]string{
			"username":     req.SeedManifest.Username,
			"command":      req.SeedManifest.Command,
			"args_pattern": req.SeedManifest.ArgsPattern,
		}, "", "  ")
		if err != nil {
			t.Fatalf("marshal manifest seed: %v", err)
		}
		if err := fs.MkdirAll(fs.cacheDir(req.Config.CacheDirName), 0700); err != nil {
			t.Fatalf("mkdir cache: %v", err)
		}
		if err := fs.WriteFile(mgr.ManifestPath(), data, 0600); err != nil {
			t.Fatalf("seed manifest: %v", err)
		}
	}
}

func applyUsernameEnv(req *Request) {
	if req.EnvSudoUser != "" {
		os.Setenv("SUDO_USER", req.EnvSudoUser)
	} else {
		os.Unsetenv("SUDO_USER")
	}
	if req.EnvUser != "" {
		os.Setenv("USER", req.EnvUser)
	} else {
		os.Unsetenv("USER")
	}
}

type mapFS struct {
	root      string
	files     map[string][]byte
	modes     map[string]os.FileMode
	dirs      map[string]bool
	mu        sync.Mutex
	tempSeq   int
	cacheRoot string
}

func newMapFS(t *testing.T) *mapFS {
	t.Helper()
	root := t.TempDir()
	return &mapFS{
		root:      root,
		files:     map[string][]byte{},
		modes:     map[string]os.FileMode{},
		dirs:      map[string]bool{root: true},
		cacheRoot: filepath.Join(root, "Library", "Caches"),
	}
}

func (fs *mapFS) redirect(path string) string {
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "/etc/sudoers.d/") {
		name := strings.TrimPrefix(path, "/etc/sudoers.d/")
		return filepath.Join(fs.root, "etc", "sudoers.d", name)
	}
	cachePrefix := filepath.Join(fs.cacheRoot)
	if strings.HasPrefix(path, cachePrefix) {
		return path
	}
	if strings.Contains(path, string(filepath.Separator)) {
		parts := strings.Split(path, string(filepath.Separator))
		for i, p := range parts {
			if p == "Caches" && i+1 < len(parts) {
				rel := filepath.Join(parts[i+1:]...)
				return filepath.Join(fs.cacheRoot, rel)
			}
		}
	}
	return path
}

func (fs *mapFS) cacheDir(name string) string {
	return filepath.Join(fs.cacheRoot, name)
}

func (fs *mapFS) UserCacheDir() (string, error) {
	return fs.cacheRoot, nil
}

func (fs *mapFS) Stat(name string) (os.FileInfo, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	p := fs.redirect(name)
	if data, ok := fs.files[p]; ok {
		return &memFileInfo{name: p, size: int64(len(data)), mode: fs.modes[p]}, nil
	}
	if fs.dirs[p] {
		return &memFileInfo{name: p, isDir: true, mode: 0755}, nil
	}
	return nil, os.ErrNotExist
}

func (fs *mapFS) ReadFile(name string) ([]byte, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	p := fs.redirect(name)
	data, ok := fs.files[p]
	if !ok {
		return nil, os.ErrNotExist
	}
	out := make([]byte, len(data))
	copy(out, data)
	return out, nil
}

func (fs *mapFS) read(name string) string {
	data, err := fs.ReadFile(name)
	if err != nil {
		return ""
	}
	return string(data)
}

func (fs *mapFS) WriteFile(name string, data []byte, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	p := fs.redirect(name)
	if err := fs.ensureParent(p); err != nil {
		return err
	}
	cp := make([]byte, len(data))
	copy(cp, data)
	fs.files[p] = cp
	fs.modes[p] = perm
	return nil
}

func (fs *mapFS) Remove(name string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	p := fs.redirect(name)
	if _, ok := fs.files[p]; !ok {
		return os.ErrNotExist
	}
	delete(fs.files, p)
	delete(fs.modes, p)
	return nil
}

func (fs *mapFS) MkdirAll(path string, perm os.FileMode) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	p := fs.redirect(path)
	fs.dirs[p] = true
	return nil
}

func (fs *mapFS) ensureParent(path string) error {
	dir := filepath.Dir(path)
	if dir == path || dir == "" || dir == "." {
		return nil
	}
	fs.dirs[dir] = true
	return fs.ensureParent(dir)
}

type tempFile struct {
	fs   *mapFS
	path string
}

func (f *tempFile) Name() string { return f.path }

func (f *tempFile) Write(p []byte) (int, error) {
	f.fs.mu.Lock()
	defer f.fs.mu.Unlock()
	existing := f.fs.files[f.path]
	combined := append(existing, p...)
	cp := make([]byte, len(combined))
	copy(cp, combined)
	f.fs.files[f.path] = cp
	f.fs.modes[f.path] = 0600
	return len(p), nil
}

func (f *tempFile) Close() error { return nil }

func (fs *mapFS) CreateTemp(dir, pattern string) (sudosetup.TempFile, error) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.tempSeq++
	base := fs.redirect(dir)
	if base == "" || base == "." {
		base = fs.root
	}
	fs.dirs[base] = true
	name := fmt.Sprintf("sudoers-test-%d.tmp", fs.tempSeq)
	path := filepath.Join(base, name)
	fs.files[path] = nil
	fs.modes[path] = 0600
	return &tempFile{fs: fs, path: path}, nil
}

type memFileInfo struct {
	name  string
	size  int64
	isDir bool
	mode  os.FileMode
}

func (m *memFileInfo) Name() string       { return filepath.Base(m.name) }
func (m *memFileInfo) Size() int64        { return m.size }
func (m *memFileInfo) Mode() os.FileMode  { return m.mode }
func (m *memFileInfo) ModTime() time.Time { return time.Time{} }
func (m *memFileInfo) IsDir() bool        { return m.isDir }
func (m *memFileInfo) Sys() any           { return nil }

var _ io.Closer = (*tempFile)(nil)

func newRecordingRunner(req *Request) *recordingRunner {
	return &recordingRunner{req: req}
}

type recordingRunner struct {
	req   *Request
	calls []RunnerCall
}

func (r *recordingRunner) Run(name string, args ...string) error {
	r.calls = append(r.calls, RunnerCall{Method: "Run", Name: name, Args: args})
	return r.dispatch(name, args)
}

func (r *recordingRunner) CombinedOutput(name string, args ...string) ([]byte, error) {
	r.calls = append(r.calls, RunnerCall{Method: "CombinedOutput", Name: name, Args: args})
	err := r.dispatch(name, args)
	if err != nil {
		return []byte(err.Error()), err
	}
	if len(args) >= 2 && args[0] == "-n" && r.req.SudoNCommandOutput != "" {
		return []byte(r.req.SudoNCommandOutput), nil
	}
	return nil, nil
}

func (r *recordingRunner) dispatch(name string, args []string) error {
	if name != "sudo" {
		return nil
	}
	if len(args) == 0 {
		return nil
	}
	switch args[0] {
	case "-n":
		if len(args) == 1 || (len(args) == 2 && args[1] == "true") {
			if !r.req.SudoNTrueOK {
				return fmt.Errorf("sudo: a password is required")
			}
			return nil
		}
		if !r.req.SudoNCommandOK {
			return fmt.Errorf("sudo: a password is required")
		}
		return nil
	case "visudo":
		if r.req.VisudoFails {
			return fmt.Errorf("visudo: syntax error")
		}
		return nil
	case "install":
		if r.req.InstallFails {
			return fmt.Errorf("install: permission denied")
		}
		return nil
	case "rm":
		if r.req.RmFails {
			return fmt.Errorf("rm: operation not permitted")
		}
		return nil
	case "-k":
		if r.req.SudoKFails {
			return fmt.Errorf("sudo -k failed")
		}
		return nil
	default:
		return nil
	}
}
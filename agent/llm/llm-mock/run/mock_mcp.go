package run

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	EnvMockMCP        = "LLM_MOCK_MCP"
	EnvMockMCPCommand = "LLM_MOCK_MCP_COMMAND"
)

var mcpSpecNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

// MCPSpec is one mock-mcp child declared via --mock-mcp or LLM_MOCK_MCP.
type MCPSpec struct {
	Name     string
	Delay    time.Duration
	DelayMin time.Duration
	DelayMax time.Duration
	Range    bool
	Hang     bool
}

func mergeMCPSpecs(flagSpecs []string, envCSV string) ([]MCPSpec, error) {
	var ordered []MCPSpec
	byName := map[string]int{}
	add := func(raw string) error {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			return nil
		}
		s, err := parseMCPSpec(raw)
		if err != nil {
			return err
		}
		if i, ok := byName[s.Name]; ok {
			ordered[i] = s
			return nil
		}
		byName[s.Name] = len(ordered)
		ordered = append(ordered, s)
		return nil
	}
	for _, raw := range strings.Split(envCSV, ",") {
		if err := add(raw); err != nil {
			return nil, err
		}
	}
	for _, raw := range flagSpecs {
		if err := add(raw); err != nil {
			return nil, err
		}
	}
	return ordered, nil
}

func parseMCPSpec(raw string) (MCPSpec, error) {
	name, rest, ok := strings.Cut(strings.TrimSpace(raw), "=")
	if !ok || strings.TrimSpace(name) == "" || strings.TrimSpace(rest) == "" {
		return MCPSpec{}, fmt.Errorf("invalid --mock-mcp spec %q: want name=duration, name=min-max, or name=hang", raw)
	}
	name = strings.TrimSpace(name)
	rest = strings.TrimSpace(rest)
	if !mcpSpecNameRe.MatchString(name) {
		return MCPSpec{}, fmt.Errorf("invalid --mock-mcp name %q: use letters, digits, _ or -", name)
	}
	s := MCPSpec{Name: name}
	if rest == "hang" {
		s.Hang = true
		return s, nil
	}
	if d, err := time.ParseDuration(rest); err == nil {
		if d < 0 {
			return MCPSpec{}, fmt.Errorf("invalid --mock-mcp spec %q: delay must be >= 0", raw)
		}
		s.Delay = d
		return s, nil
	}
	minD, maxD, err := parseDurationRange(rest)
	if err != nil {
		return MCPSpec{}, fmt.Errorf("invalid --mock-mcp spec %q: %w", raw, err)
	}
	s.DelayMin = minD
	s.DelayMax = maxD
	s.Range = true
	return s, nil
}

func parseDurationRange(rest string) (time.Duration, time.Duration, error) {
	for i := strings.LastIndex(rest, "-"); i > 0; i = strings.LastIndex(rest[:i], "-") {
		left, right := rest[:i], rest[i+1:]
		minD, err1 := time.ParseDuration(left)
		maxD, err2 := time.ParseDuration(right)
		if err1 == nil && err2 == nil {
			if minD < 0 || maxD < 0 {
				return 0, 0, fmt.Errorf("delay must be >= 0")
			}
			if minD > maxD {
				return 0, 0, fmt.Errorf("delay-min (%s) > delay-max (%s)", minD, maxD)
			}
			return minD, maxD, nil
		}
	}
	return 0, 0, fmt.Errorf("want duration, min-max, or hang")
}

func (s MCPSpec) args() []string {
	out := []string{"--name", s.Name}
	if s.Hang {
		return append(out, "--hang")
	}
	if s.Range {
		return append(out, "--delay-min", s.DelayMin.String(), "--delay-max", s.DelayMax.String())
	}
	if s.Delay > 0 {
		return append(out, "--delay", s.Delay.String())
	}
	return out
}

func (s MCPSpec) startupTimeoutSec() int {
	if s.Hang {
		return 600
	}
	max := s.Delay
	if s.DelayMax > max {
		max = s.DelayMax
	}
	if s.DelayMin > max {
		max = s.DelayMin
	}
	sec := int(math.Ceil(max.Seconds())) + 30
	if sec < 30 {
		sec = 30
	}
	return sec
}

func formatMCPBlocks(bin string, specs []MCPSpec) string {
	if len(specs) == 0 {
		return ""
	}
	var b strings.Builder
	for _, s := range specs {
		fmt.Fprintf(&b, "\n[mcp_servers.%s]\n", s.Name)
		fmt.Fprintf(&b, "command = %s\n", strconv.Quote(bin))
		fmt.Fprintf(&b, "args = %s\n", tomlStringArray(s.args()))
		fmt.Fprintf(&b, "startup_timeout_sec = %d\n", s.startupTimeoutSec())
	}
	return b.String()
}

func tomlStringArray(ss []string) string {
	if len(ss) == 0 {
		return "[]"
	}
	parts := make([]string, len(ss))
	for i, s := range ss {
		parts[i] = strconv.Quote(s)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func resolveMockMCPPath() (string, error) {
	if v := strings.TrimSpace(os.Getenv(EnvMockMCPCommand)); v != "" {
		return v, nil
	}
	if exe, err := os.Executable(); err == nil {
		sibling := filepath.Join(filepath.Dir(exe), "mock-mcp")
		if st, err := os.Stat(sibling); err == nil && !st.IsDir() {
			if abs, err := filepath.Abs(sibling); err == nil {
				return abs, nil
			}
			return sibling, nil
		}
	}
	p, err := exec.LookPath("mock-mcp")
	if err != nil {
		return "", fmt.Errorf("mock-mcp not found (go install github.com/xhd2015/agent-pro/agent/llm/mock-mcp or set %s)", EnvMockMCPCommand)
	}
	return p, nil
}

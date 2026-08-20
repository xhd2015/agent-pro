package run

import (
	"strings"
	"testing"
	"time"
)

func TestParseMCPSpec_range(t *testing.T) {
	s, err := parseMCPSpec("slow_01=1s-10s")
	if err != nil {
		t.Fatal(err)
	}
	if s.Name != "slow_01" || !s.Range || s.DelayMin != time.Second || s.DelayMax != 10*time.Second {
		t.Fatalf("%+v", s)
	}
}

func TestParseMCPSpec_fixedAndHang(t *testing.T) {
	fixed, err := parseMCPSpec("ready=5s")
	if err != nil {
		t.Fatal(err)
	}
	if fixed.Delay != 5*time.Second || fixed.Range || fixed.Hang {
		t.Fatalf("%+v", fixed)
	}
	hang, err := parseMCPSpec("slow_hang=hang")
	if err != nil {
		t.Fatal(err)
	}
	if !hang.Hang {
		t.Fatalf("%+v", hang)
	}
}

func TestMergeMCPSpecs_cliWinsDuplicateName(t *testing.T) {
	specs, err := mergeMCPSpecs([]string{"slow_01=5s"}, "slow_01=1s-10s,slow_02=10s-30s")
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 2 {
		t.Fatalf("len=%d", len(specs))
	}
	if specs[0].Name != "slow_01" || specs[0].Range || specs[0].Delay != 5*time.Second {
		t.Fatalf("cli should win: %+v", specs[0])
	}
	if specs[1].Name != "slow_02" || !specs[1].Range {
		t.Fatalf("env second kept: %+v", specs[1])
	}
}

func TestFormatMCPBlocks_startupTimeout(t *testing.T) {
	s, err := parseMCPSpec("slow_04=60s-120s")
	if err != nil {
		t.Fatal(err)
	}
	body := formatMCPBlocks("/tmp/mock-mcp", []MCPSpec{s})
	if !strings.Contains(body, "[mcp_servers.slow_04]") {
		t.Fatalf("missing table:\n%s", body)
	}
	if !strings.Contains(body, `command = "/tmp/mock-mcp"`) {
		t.Fatalf("command:\n%s", body)
	}
	if !strings.Contains(body, `"--delay-min"`) || !strings.Contains(body, `"--delay-max"`) {
		t.Fatalf("args:\n%s", body)
	}
	if !strings.Contains(body, "startup_timeout_sec = 150") {
		t.Fatalf("timeout:\n%s", body)
	}
}

func TestParseMCPSpec_minGreaterMax(t *testing.T) {
	_, err := parseMCPSpec("bad=10s-1s")
	if err == nil {
		t.Fatal("expected error")
	}
}

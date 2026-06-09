package agentui

import (
	"testing"
)

func TestWrapText(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		width  int
		expect string
	}{
		{
			name:   "no wrap needed",
			input:  "hello world",
			width:  20,
			expect: "hello world",
		},
		{
			name:   "wrap at word boundary",
			input:  "hello world here",
			width:  11,
			expect: "hello world\nhere",
		},
		{
			name:   "multiline input",
			input:  "line one\nline two long",
			width:  8,
			expect: "line one\nline two\nlong",
		},
		{
			name:   "long word forced break",
			input:  "supercalifragilistic",
			width:  10,
			expect: "supercalif\nragilistic",
		},
		{
			name:   "zero width passes through",
			input:  "hello world",
			width:  0,
			expect: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WrapText(tt.input, tt.width)
			if got != tt.expect {
				t.Errorf("WrapText(%q, %d) = %q, want %q", tt.input, tt.width, got, tt.expect)
			}
		})
	}
}

func TestConfigRunHonorsAgentName(t *testing.T) {
	cfg := Config{
		AgentName:     "my-agent",
		SessionPrefix: "ma_",
		Prompt:        "prompt",
		Usage:         "my usage",
	}
	if cfg.AgentName != "my-agent" {
		t.Errorf("expected AgentName 'my-agent', got %q", cfg.AgentName)
	}
	if cfg.SessionPrefix != "ma_" {
		t.Errorf("expected SessionPrefix 'ma_', got %q", cfg.SessionPrefix)
	}
}

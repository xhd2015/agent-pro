package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFormatTextMarksDefault(t *testing.T) {
	t.Parallel()
	cat := Catalog{
		Home:    "/tmp/.grok",
		Default: "grok-4.5",
		Models:  []string{"ais-glm-5-2", "grok-4.5", "grok-4.6"},
	}
	out := FormatText(cat)
	if !strings.Contains(out, "* grok-4.5\n") {
		t.Fatalf("missing default mark:\n%s", out)
	}
	if !strings.Contains(out, "  ais-glm-5-2\n") || !strings.Contains(out, "  grok-4.6\n") {
		t.Fatalf("missing indented models:\n%s", out)
	}
	if strings.Contains(out, "* ais-glm-5-2") {
		t.Fatalf("non-default marked:\n%s", out)
	}
}

func TestFormatTextEmpty(t *testing.T) {
	t.Parallel()
	out := FormatText(Catalog{Home: "/tmp/.grok", Models: []string{}})
	if !strings.Contains(out, "(no models)") {
		t.Fatalf("output=%q", out)
	}
}

func TestFormatJSONRoundTrip(t *testing.T) {
	t.Parallel()
	cat := Catalog{
		Home:       "/tmp/.grok",
		Default:    "grok-4.5",
		Models:     []string{"grok-4.5"},
		FromConfig: true,
		FromCache:  true,
	}
	raw, err := FormatJSON(cat)
	if err != nil {
		t.Fatal(err)
	}
	var got Catalog
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Home != cat.Home || got.Default != cat.Default || len(got.Models) != 1 || got.Models[0] != "grok-4.5" {
		t.Fatalf("got=%+v", got)
	}
}

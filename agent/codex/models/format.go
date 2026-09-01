package models

import (
	"encoding/json"
	"fmt"
	"strings"
)

// FormatJSON returns indented Catalog JSON for CLI --json output.
func FormatJSON(cat Catalog) ([]byte, error) {
	return json.MarshalIndent(cat, "", "  ")
}

// FormatText returns a human-readable catalog listing.
// The configured default model is marked with a leading "* ".
func FormatText(cat Catalog) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Home: %s\n", cat.Home)
	if cat.Default != "" {
		fmt.Fprintf(&b, "Default: %s\n", cat.Default)
	} else {
		b.WriteString("Default:\n")
	}
	b.WriteByte('\n')
	if len(cat.Models) == 0 {
		b.WriteString("(no models)\n")
		return b.String()
	}
	for _, m := range cat.Models {
		mark := "  "
		if cat.Default != "" && m.Slug == cat.Default {
			mark = "* "
		}
		line := mark + m.Slug
		if m.DisplayName != "" {
			line += "  " + m.DisplayName
		}
		if len(m.Reasoning) > 0 {
			line += "  reasoning=[" + strings.Join(m.Reasoning, " ") + "]"
		}
		if m.DefaultReasoning != "" {
			line += "  default=" + m.DefaultReasoning
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

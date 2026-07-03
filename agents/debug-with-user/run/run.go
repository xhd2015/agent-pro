// Package run exposes embedded skill content for agent-pro registration.
package run

import _ "embed"

//go:embed SKILL.md
var SkillFile string
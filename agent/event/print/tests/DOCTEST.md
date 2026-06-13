# Print Package Tests

These doc-style tests verify the `agent/event/print` package. The package
parses agent event JSONL lines and formats them as compact human-readable
strings.

Tests import the print package and the opencode adapter (for side-effect
adapter registration) and call `FormatTraceLine` with various JSONL event
lines.

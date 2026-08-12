# Scenario

**Feature**: `agent-run run` documents and registers `--prompt-file`

```
# help surface (pure file read of run_cmd.go — no process stdio)
runHelp / run flag registration -> documents --prompt-file

# flag wire
scan pkgs/agentruncli/*.go -> "--prompt-file" literal present
```

## Steps

1. Leaves set mode (`cli_help` or `source_wire`).
2. Help leaf reads `pkgs/agentruncli/run_cmd.go` (parallel-safe; no Handle).
3. Source-wire leaf scans production `.go` under `pkgs/agentruncli`.

Organization-only grouping node (no Setup Go block).

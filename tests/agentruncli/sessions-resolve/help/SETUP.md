# Scenario

**Feature**: help text for resolve verb and sessions index

```
RunSessions(["resolve", "--help"]) -> nil; documents Usage + flags
RunSessions(["-h"]) -> nil; sessions help mentions resolve
```

## Preconditions

- Help returns nil from `RunSessions` (flags.Help path); no store lookup required.
- Grouping for help-only leaves (no session seeds).

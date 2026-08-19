#!/usr/bin/env python3
"""Migrate doctest assert templates from version: 2 to version: 3.

v2: content lines are literal unless hasRegexIntent (then raw RE).
v3: every content line is raw RE — escape former literals; leave regex-intent
lines and omit markers unchanged. Preserve __PLACEHOLDER__ and <ansi-color> tags.

Also handles:
- same-line closing backtick for Go raw strings
- strings.Join double-quoted bodies (double RE backslashes in Go source)
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

_META = frozenset(r"\.+*?()|[]{}^$")
_PH = re.compile(r"__[A-Z][A-Z0-9_]*__")
_OMIT = re.compile(r"^\.\.\.\s*\d+\s*lines\s+omitted\s*\.\.\.$")
_ANSI_OPEN = "<ansi-color"
_ANSI_CLOSE = "</ansi-color>"

# Body ends at markdown fence ``` or Go raw-string ` (newline optional before `).
_BLOCK = re.compile(
    r"---\nversion: 2\n(.*?)---\n(.*?)(?=\n```|```|\n`|`)",
    re.DOTALL,
)


def quote_meta(s: str) -> str:
    return "".join("\\" + ch if ch in _META else ch for ch in s)


def mask_protected(line: str) -> str:
    out = line
    for name in _PH.findall(line):
        out = out.replace(name, " " * len(name))
    while True:
        start = out.find(_ANSI_OPEN)
        if start < 0:
            break
        open_end = out.find(">", start)
        if open_end < 0:
            break
        close_start = out.find(_ANSI_CLOSE, open_end + 1)
        if close_start < 0:
            break
        span_end = close_start + len(_ANSI_CLOSE)
        out = out[:start] + (" " * (span_end - start)) + out[span_end:]
    return out


def find_balanced_bracket(s: str, start: int) -> bool:
    if start >= len(s) or s[start] != "[":
        return False
    depth = 0
    for i in range(start, len(s)):
        if s[i] == "[":
            depth += 1
        elif s[i] == "]":
            depth -= 1
            if depth == 0:
                return True
    return False


def is_alternation(s: str, pipe_idx: int) -> bool:
    return bool(s[:pipe_idx].strip() and s[pipe_idx + 1 :].strip())


def is_quantifier_brace(s: str, start: int) -> bool:
    if start >= len(s) or s[start] != "{":
        return False
    end = s.find("}", start)
    if end < 0:
        return False
    inner = s[start + 1 : end]
    if not inner:
        return False
    for ch in inner:
        if ch == ",":
            continue
        if ch < "0" or ch > "9":
            return False
    return True


def scan_regex_signals(s: str) -> bool:
    i, n = 0, len(s)
    while i < n:
        c = s[i]
        if c == ".":
            if i + 1 < n and s[i + 1] in "*+?.":
                return True
        elif c == "^" and i == 0:
            return True
        elif c == "$" and i == n - 1:
            return True
        elif c == "\\" and i + 1 < n and s[i + 1] in "dDwWsSbB":
            return True
        elif c == "[" and find_balanced_bracket(s, i):
            return True
        elif c == "|" and is_alternation(s, i):
            return True
        elif c == "{" and is_quantifier_brace(s, i):
            return True
        elif c in "*+" and i + 1 < n and s[i + 1] == "?":
            return True
        i += 1
    return False


def has_regex_intent(line: str) -> bool:
    return scan_regex_signals(mask_protected(line))


def escape_literal_line(line: str) -> str:
    out: list[str] = []
    i, n = 0, len(line)
    while i < n:
        if line.startswith(_ANSI_OPEN, i):
            gt = line.find(">", i)
            if gt < 0:
                out.append(quote_meta(line[i:]))
                break
            close = line.find(_ANSI_CLOSE, gt)
            if close < 0:
                out.append(quote_meta(line[i:]))
                break
            out.append(line[i : close + len(_ANSI_CLOSE)])
            i = close + len(_ANSI_CLOSE)
            continue
        m = _PH.match(line, i)
        if m:
            out.append(m.group(0))
            i = m.end()
            continue
        j = i + 1
        while j < n:
            if line.startswith(_ANSI_OPEN, j) or _PH.match(line, j):
                break
            j += 1
        out.append(quote_meta(line[i:j]))
        i = j
    return "".join(out)


def transform_content_line(line: str) -> str:
    if not line:
        return line
    if _OMIT.match(line.strip()):
        return line
    if has_regex_intent(line):
        return line
    return escape_literal_line(line)


def transform_body(body: str) -> str:
    ends_nl = body.endswith("\n")
    lines = body.split("\n")
    if ends_nl and lines and lines[-1] == "":
        return "\n".join(transform_content_line(l) for l in lines[:-1]) + "\n"
    return "\n".join(transform_content_line(l) for l in lines)


def transform_block(m: re.Match[str]) -> str:
    header_rest = m.group(1)
    body = m.group(2)
    return f"---\nversion: 3\n{header_rest}---\n{transform_body(body)}"


def transform_join_strings(text: str) -> str:
    lines = text.splitlines(keepends=True)
    out: list[str] = []
    in_join_body = False
    for line in lines:
        if re.search(r'"version:\s*3"', line):
            in_join_body = True
            out.append(line)
            continue
        if in_join_body:
            if re.match(r"\s*\},?\s*$", line) or '}, "\\n")' in line or '}, "\\n")' in line:
                in_join_body = False
                out.append(line)
                continue
            m = re.match(r'^(\s*")(.*)("(?:\s*\+\s*full)?\s*,?\s*)$', line.rstrip("\n"))
            nl = "\n" if line.endswith("\n") else ""
            if m:
                prefix, content, suffix = m.group(1), m.group(2), m.group(3)
                if content in ("---",) or content.startswith("version:"):
                    out.append(line)
                    continue
                transformed = transform_content_line(content)
                # Go double-quoted source: each RE backslash must be written twice.
                transformed = transformed.replace("\\", "\\\\")
                out.append(f"{prefix}{transformed}{suffix}{nl}")
                continue
            if re.match(r'^\s*"",?\s*$', line):
                out.append(line)
                continue
            if not line.strip().startswith('"'):
                in_join_body = False
        out.append(line)
    return "".join(out)


def transform_file(text: str) -> tuple[str, int]:
    new_text, n = _BLOCK.subn(transform_block, text)
    new_text = re.sub(r"version:\s*2\b", "version: 3", new_text)
    if "strings.Join" in new_text:
        new_text = transform_join_strings(new_text)
    return new_text, n


def main() -> int:
    root = Path(sys.argv[1] if len(sys.argv) > 1 else ".")
    files = sorted(
        p
        for p in root.rglob("ASSERT.md")
        if "version: 2" in p.read_text(encoding="utf-8", errors="replace")
    )
    if not files:
        print("no version: 2 ASSERT.md files found", file=sys.stderr)
        return 1
    changed = 0
    for path in files:
        text = path.read_text(encoding="utf-8")
        new_text, nblocks = transform_file(text)
        if new_text != text:
            path.write_text(new_text, encoding="utf-8")
            changed += 1
            print(f"updated {path} ({nblocks} yaml blocks)")
    print(f"done: {changed}/{len(files)} files")
    left = [
        str(p)
        for p in root.rglob("ASSERT.md")
        if "version: 2" in p.read_text(encoding="utf-8", errors="replace")
    ]
    if left:
        print("WARNING residual version: 2:", *left, sep="\n  ", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    sys.exit(main())

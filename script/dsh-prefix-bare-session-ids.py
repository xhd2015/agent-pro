#!/usr/bin/env python3
"""Rename bare-UUID DSH session dirs to session-<uuid> and match the log header id.

Dry-run by default. Apply with --apply.

  DSH_HOME=~/.dsh python3 script/dsh-prefix-bare-session-ids.py
  DSH_HOME=~/.dsh python3 script/dsh-prefix-bare-session-ids.py --apply
"""

from __future__ import annotations

import argparse
import fcntl
import json
import os
import re
import subprocess
import sys
from pathlib import Path

UUID = re.compile(
    r"^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$",
    re.I,
)
LOG = re.compile(r"^session(?:\.v[1-9][0-9]*)?\.jsonl(\.zstd)?$")


def sessions_root() -> Path:
    home = os.environ.get("DSH_HOME", "").strip()
    if not home:
        home = str(Path.home() / ".dsh")
    return Path(home) / "sessions"


def rewrite_header(path: Path, old_id: str, new_id: str, apply: bool) -> bool:
    zstd = path.suffix == ".zstd"
    if zstd:
        raw = subprocess.check_output(["zstdcat", str(path)])
    else:
        raw = path.read_bytes()
    lines = raw.split(b"\n", 1)
    if not lines or not lines[0].strip():
        return False
    header = json.loads(lines[0])
    if header.get("type") != "session" or header.get("id") != old_id:
        return False
    if not apply:
        return True
    header["id"] = new_id
    out = json.dumps(header, separators=(",", ":"), ensure_ascii=False).encode() + b"\n"
    if len(lines) > 1:
        out += lines[1]
    tmp = path.with_name(path.name + ".tmp")
    if zstd:
        proc = subprocess.run(["zstd", "-q", "-o", str(tmp)], input=out, check=True)
        _ = proc
    else:
        tmp.write_bytes(out)
    os.replace(tmp, path)
    return True


def locked(lock: Path) -> bool:
    if not lock.exists():
        return False
    fd = os.open(lock, os.O_RDWR)
    try:
        fcntl.flock(fd, fcntl.LOCK_EX | fcntl.LOCK_NB)
        fcntl.flock(fd, fcntl.LOCK_UN)
        return False
    except BlockingIOError:
        return True
    finally:
        os.close(fd)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--apply", action="store_true", help="rename dirs and rewrite headers")
    args = parser.parse_args()
    root = sessions_root()
    if not root.is_dir():
        print(f"no sessions root: {root}", file=sys.stderr)
        return 1
    planned = 0
    skipped = 0
    for project in sorted(root.iterdir()):
        if not project.is_dir():
            continue
        for src in sorted(project.iterdir()):
            if not src.is_dir() or not UUID.match(src.name):
                continue
            dest = src.with_name("session-" + src.name)
            if dest.exists():
                print(f"skip exists {dest}")
                skipped += 1
                continue
            if locked(src / "session.lock"):
                print(f"skip live {src}")
                skipped += 1
                continue
            logs = [p for p in src.iterdir() if LOG.match(p.name)]
            print(f"{src.name} -> {dest.name}")
            planned += 1
            if not args.apply:
                continue
            for log in logs:
                rewrite_header(log, src.name, dest.name, True)
            src.rename(dest)
    print(f"{'applied' if args.apply else 'dry-run'} {planned} skip {skipped}")
    return 0


if __name__ == "__main__":
    sys.exit(main())

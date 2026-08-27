#!/usr/bin/env bash
# Emit doctest path args for one shard of a test root's immediate child dirs.
#
# Usage:
#   script/ci/doctest-shard-paths.sh --root DIR --shard I --shards N
#
# Prints space-separated paths like:
#   ./cmd/agent-run/tests/foo/... ./cmd/agent-run/tests/bar/...
#
# Sharding: sort child directory names, keep index i where i % N == I.
# New packages under ROOT are included automatically.
#
# Exit 0 with empty stdout if this shard has no dirs (CI should skip doctest).

set -euo pipefail

root=""
shard=""
shards=""

usage() {
  echo "usage: $0 --root DIR --shard I --shards N" >&2
  exit 2
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --root)
      root="${2:-}"
      shift 2
      ;;
    --shard)
      shard="${2:-}"
      shift 2
      ;;
    --shards)
      shards="${2:-}"
      shift 2
      ;;
    -h | --help)
      usage
      ;;
    *)
      echo "unknown arg: $1" >&2
      usage
      ;;
  esac
done

if [[ -z "$root" || -z "$shard" || -z "$shards" ]]; then
  usage
fi
if ! [[ "$shard" =~ ^[0-9]+$ && "$shards" =~ ^[0-9]+$ && "$shards" -gt 0 ]]; then
  echo "shard and shards must be non-negative integers; shards > 0" >&2
  exit 2
fi
if [[ "$shard" -ge "$shards" ]]; then
  echo "shard ($shard) must be < shards ($shards)" >&2
  exit 2
fi

root="${root#./}"
root="${root%/}"
if [[ ! -d "$root" ]]; then
  echo "root is not a directory: $root" >&2
  exit 1
fi

# Collect immediate child directories (sorted for stable shard assignment).
# Portable across GNU/BSD find (no -printf).
mapfile -t dirs < <(
  find "$root" -mindepth 1 -maxdepth 1 -type d | while IFS= read -r d; do
    basename "$d"
  done | LC_ALL=C sort
)

paths=()
for i in "${!dirs[@]}"; do
  if (( i % shards == shard )); then
    paths+=("./${root}/${dirs[$i]}/...")
  fi
done

if [[ ${#paths[@]} -eq 0 ]]; then
  exit 0
fi

printf '%s' "${paths[0]}"
for ((i = 1; i < ${#paths[@]}; i++)); do
  printf ' %s' "${paths[$i]}"
done
printf '\n'

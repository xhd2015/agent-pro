#!/usr/bin/env bash
# stress-git-index-race.sh — race concurrent git index writers until:
#   fatal: unable to write new index file
# (optionally also related index.lock failures)
#
# Isolates work in a throwaway repo under /tmp (never touches real checkouts).
#
# Usage:
#   ./script/debug/stress-git-index-race.sh
#   ./script/debug/stress-git-index-race.sh --max 5000 --workers 6 --mode commit-add
#   ./script/debug/stress-git-index-race.sh --mode all --files 237 --linked-worktree
#   ./script/debug/stress-git-index-race.sh --strict --mode kill-mid --max 20000
#
# Exit codes:
#   0  bug reproduced (exact, or related if not --strict)
#   1  exhausted --max without acceptable hit
#   2  harness / setup error

set -euo pipefail

MAX=10000
WORKERS=4
MODE="commit-add"
FILES=200
KEEP_ON_HIT=1
LINKED_WORKTREE=0
WORKDIR=""
OPS_PER_WORKER=40
VERBOSE=0
# accept-related (default): exact OR related lock errors count as hit
# strict: only "unable to write new index file"
# until-exact: log related hits but keep going until exact message
MATCH_POLICY="accept-related"

EXACT_PATTERN='unable to write new index file'
RELATED_PATTERNS=(
  'Unable to create .*index\.lock'
  'index\.lock.*File exists'
  'Unable to write index file'
)

usage() {
  cat <<'EOF'
Usage: stress-git-index-race.sh [options]

Stress concurrent git index writers to reproduce index write/lock failures.

Primary target (exact):
  fatal: unable to write new index file

Related (lock contention — common under concurrent writers):
  fatal: Unable to create '.../index.lock': File exists.

Options:
  --max N              Max rounds (default: 10000)
  --workers W          Concurrent workers per round (default: 4)
  --mode MODE          commit-add | commit-commit | commit-refresh |
                       hook-rewrite | kill-mid | immutable-index |
                       exact-control | all (default: commit-add)
                       exact-control = deterministic exact message (macOS chflags)
  --files F            Seed file count to inflate index rewrites (default: 200)
  --ops N              Ops each worker runs per round (default: 40)
  --linked-worktree    Use a linked worktree (checkout path != git common dir)
  --workdir PATH       Fixed work root (default: mktemp under /tmp)
  --strict             Only count exact "unable to write new index file"
  --until-exact        Keep running after related hits until exact message
  --no-keep-on-hit     Remove workdir even after a hit
  --verbose            Print per-round progress
  -h, --help           Show this help

Exit: 0 = reproduced, 1 = not reproduced, 2 = setup error

Notes:
  - Concurrent git almost always hits related index.lock errors first.
  - The exact message usually means the process held the lock but write/rename
    of the index failed (I/O error, kill mid-write, FS glitch). Use
    --mode kill-mid --strict to push toward that path.
EOF
}

log() { printf '%s\n' "$*"; }
err() { printf 'Error: %s\n' "$*" >&2; }
vlog() { if (( VERBOSE )); then printf '  %s\n' "$*"; fi; }

while [[ $# -gt 0 ]]; do
  case "$1" in
    --max) MAX="${2:?}"; shift 2 ;;
    --workers) WORKERS="${2:?}"; shift 2 ;;
    --mode) MODE="${2:?}"; shift 2 ;;
    --files) FILES="${2:?}"; shift 2 ;;
    --ops) OPS_PER_WORKER="${2:?}"; shift 2 ;;
    --linked-worktree) LINKED_WORKTREE=1; shift ;;
    --workdir) WORKDIR="${2:?}"; shift 2 ;;
    --strict) MATCH_POLICY="strict"; shift ;;
    --until-exact) MATCH_POLICY="until-exact"; shift ;;
    --no-keep-on-hit) KEEP_ON_HIT=0; shift ;;
    --verbose|-v) VERBOSE=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) err "unknown flag: $1"; usage >&2; exit 2 ;;
  esac
done

if ! [[ "$MAX" =~ ^[0-9]+$ ]] || (( MAX < 1 )); then
  err "--max must be a positive integer"; exit 2
fi
if ! [[ "$WORKERS" =~ ^[0-9]+$ ]] || (( WORKERS < 2 )); then
  err "--workers must be >= 2"; exit 2
fi
if ! [[ "$FILES" =~ ^[0-9]+$ ]] || (( FILES < 1 )); then
  err "--files must be a positive integer"; exit 2
fi

case "$MODE" in
  commit-add|commit-commit|commit-refresh|hook-rewrite|kill-mid|immutable-index|exact-control|all) ;;
  *) err "unknown --mode: $MODE"; usage >&2; exit 2 ;;
esac

if [[ -z "$WORKDIR" ]]; then
  WORKDIR="$(mktemp -d /tmp/git-index-race.XXXXXX)"
else
  mkdir -p "$WORKDIR"
fi

REPO_MAIN="$WORKDIR/main"
HIT_DIR="$WORKDIR/hit"
ROUNDS_DIR="$WORKDIR/rounds"
mkdir -p "$REPO_MAIN" "$HIT_DIR" "$ROUNDS_DIR"

WORKER_PIDS=()
cleanup_workers() {
  local p
  for p in "${WORKER_PIDS[@]:-}"; do
    kill "$p" 2>/dev/null || true
    # also kill process groups if workers spawned children
    kill -- -"$p" 2>/dev/null || true
  done
  wait 2>/dev/null || true
  WORKER_PIDS=()
}

trap 'cleanup_workers' EXIT

export GIT_CONFIG_GLOBAL=/dev/null
export GIT_CONFIG_SYSTEM=/dev/null
export GIT_CONFIG_NOSYSTEM=1
unset GIT_DIR GIT_WORK_TREE GIT_COMMON_DIR 2>/dev/null || true

git_env() {
  env \
    GIT_AUTHOR_NAME='race-bot' \
    GIT_AUTHOR_EMAIL='race-bot@example.com' \
    GIT_COMMITTER_NAME='race-bot' \
    GIT_COMMITTER_EMAIL='race-bot@example.com' \
    "$@"
}

# Always absolute — relative ".git" would chflags/rm the caller's cwd repo.
abs_git_dir() {
  local repo="$1"
  local gdir
  gdir="$(git_env git -C "$repo" rev-parse --absolute-git-dir 2>/dev/null)" \
    || gdir="$(git_env git -C "$repo" rev-parse --git-dir)"
  if [[ "$gdir" != /* ]]; then
    gdir="$repo/$gdir"
  fi
  printf '%s\n' "$gdir"
}

setup_repo() {
  local main="$REPO_MAIN"
  rm -rf "$main"
  mkdir -p "$main"
  git_env git -C "$main" init -b main >/dev/null
  # Disable hooks by default (re-enabled for hook-rewrite).
  git_env git -C "$main" config core.hooksPath /dev/null
  git_env git -C "$main" config user.name 'race-bot'
  git_env git -C "$main" config user.email 'race-bot@example.com'

  local i
  mkdir -p "$main/files"
  for (( i = 1; i <= FILES; i++ )); do
    printf 'seed %s line %s\n' "$i" "0" >"$main/files/f-$(printf '%04d' "$i").txt"
  done
  printf '# race harness\n' >"$main/README.md"
  git_env git -C "$main" add -A
  git_env git -C "$main" commit -m 'seed' >/dev/null

  if (( LINKED_WORKTREE )); then
    local wt="$WORKDIR/worktree"
    rm -rf "$wt"
    git_env git -C "$main" worktree add "$wt" -b race-wt main >/dev/null 2>&1 \
      || git_env git -C "$main" worktree add "$wt" main >/dev/null
    ACTIVE_REPO="$wt"
  else
    ACTIVE_REPO="$main"
  fi

  if [[ "$MODE" == "hook-rewrite" || "$MODE" == "all" ]]; then
    install_hook_rewrite
  fi
}

install_hook_rewrite() {
  local common hooks
  common="$(git_env git -C "$ACTIVE_REPO" rev-parse --git-common-dir)"
  hooks="$common/hooks"
  mkdir -p "$hooks"
  cat >"$hooks/pre-commit" <<'HOOK'
#!/usr/bin/env bash
# Deliberately churn the index mid-commit (auto-stage style).
set -euo pipefail
root="$(git rev-parse --show-toplevel)"
f="$(git ls-files | head -1 || true)"
if [[ -n "$f" && -f "$root/$f" ]]; then
  echo "hook-touch $(date +%s%N)" >>"$root/$f"
  git add -- "$f" 2>/dev/null || true
fi
exit 0
HOOK
  chmod +x "$hooks/pre-commit"
  git_env git -C "$ACTIVE_REPO" config --unset-all core.hooksPath 2>/dev/null || true
}

reset_round() {
  local gdir
  gdir="$(abs_git_dir "$ACTIVE_REPO")"
  # Clear any leftover immutable flags from immutable-index mode.
  if [[ "$(uname -s)" == "Darwin" ]]; then
    chflags nouchg "$gdir/index" 2>/dev/null || true
    chflags nouchg "$gdir/index.lock" 2>/dev/null || true
  fi
  rm -f "$gdir/index.lock" 2>/dev/null || true
  git_env git -C "$ACTIVE_REPO" reset --hard HEAD >/dev/null 2>&1 || true
  git_env git -C "$ACTIVE_REPO" clean -fd >/dev/null 2>&1 || true

  local i f
  for (( i = 1; i <= FILES; i++ )); do
    f="$ACTIVE_REPO/files/f-$(printf '%04d' "$i").txt"
    if [[ -f "$f" ]]; then
      printf 'seed %s line %s\n' "$i" "$RANDOM" >"$f"
    fi
  done
  git_env git -C "$ACTIVE_REPO" add -A >/dev/null 2>&1 || true
}

classify_log() {
  # Sets HIT_KIND to exact|related|none and HIT_LOG to the matching file.
  local logf="$1"
  if grep -E -q "$EXACT_PATTERN" "$logf" 2>/dev/null; then
    HIT_KIND="exact"
    HIT_LOG="$logf"
    return 0
  fi
  local p
  for p in "${RELATED_PATTERNS[@]}"; do
    if grep -E -q "$p" "$logf" 2>/dev/null; then
      HIT_KIND="related"
      HIT_LOG="$logf"
      return 0
    fi
  done
  HIT_KIND="none"
  return 1
}

scan_round_logs() {
  local round_dir="$1"
  local f best_kind="none" best_log=""
  shopt -s nullglob
  for f in "$round_dir"/*.log; do
    if classify_log "$f"; then
      if [[ "$HIT_KIND" == "exact" ]]; then
        return 0
      fi
      if [[ "$best_kind" != "exact" ]]; then
        best_kind="$HIT_KIND"
        best_log="$HIT_LOG"
      fi
    fi
  done
  shopt -u nullglob
  if [[ "$best_kind" != "none" ]]; then
    HIT_KIND="$best_kind"
    HIT_LOG="$best_log"
    return 0
  fi
  HIT_KIND="none"
  return 1
}

worker_commit_add() {
  local repo="$1" logf="$2" id="$3" n="$4"
  local i f
  for (( i = 0; i < n; i++ )); do
    f="files/f-$(printf '%04d' "$(( (RANDOM % FILES) + 1 ))").txt"
    echo "w${id}-$i-$RANDOM" >>"$repo/$f" 2>/dev/null || true
    {
      git_env git -C "$repo" add -A
      git_env git -C "$repo" commit -m "w${id}-$i" --allow-empty
    } >>"$logf" 2>&1 || true
  done
}

worker_commit_only() {
  local repo="$1" logf="$2" id="$3" n="$4"
  local i
  for (( i = 0; i < n; i++ )); do
    {
      git_env git -C "$repo" commit -m "empty-w${id}-$i" --allow-empty
    } >>"$logf" 2>&1 || true
  done
}

worker_refresh() {
  local repo="$1" logf="$2" id="$3" n="$4"
  local i
  for (( i = 0; i < n; i++ )); do
    {
      git_env git -C "$repo" update-index --refresh
      git_env git -C "$repo" status -sb
      git_env git -C "$repo" add -A
    } >>"$logf" 2>&1 || true
  done
}

worker_add_churn() {
  local repo="$1" logf="$2" id="$3" n="$4"
  local i f
  for (( i = 0; i < n; i++ )); do
    f="files/f-$(printf '%04d' "$(( (RANDOM % FILES) + 1 ))").txt"
    echo "add-churn w${id} $i $RANDOM" >>"$repo/$f" 2>/dev/null || true
    {
      git_env git -C "$repo" add -A
      git_env git -C "$repo" update-index --refresh
    } >>"$logf" 2>&1 || true
  done
}

# On macOS, hold UF_IMMUTABLE on the index while peer writers run.
# When git holds index.lock and renames onto an immutable index, you get:
#   fatal: unable to write new index file
# (verified: chflags uchg .git/index && git commit)
worker_immutable_flip() {
  local repo="$1" logf="$2" id="$3" n="$4"
  local i gdir index hold_ms
  gdir="$(abs_git_dir "$repo")"
  index="$gdir/index"
  if [[ "$(uname -s)" != "Darwin" ]]; then
    echo "immutable-index mode requires macOS chflags; falling back to add-churn" >>"$logf"
    worker_add_churn "$repo" "$logf" "$id" "$n"
    return 0
  fi
  for (( i = 0; i < n; i++ )); do
    if [[ ! -f "$index" ]]; then
      sleep 0.01
      continue
    fi
    # Hold long enough that concurrent commit/add hit the rename path.
    hold_ms=$(( 20 + (RANDOM % 80) ))
    {
      chflags uchg "$index" || true
      # Busy-wait in 1ms steps (portable enough on macOS bash).
      local t=0
      while (( t < hold_ms )); do
        sleep 0.001
        t=$((t + 1))
      done
      chflags nouchg "$index" || true
      # Clear stray lock from a killed/failed peer so the round can continue.
      if [[ -f "${index}.lock" ]]; then
        chflags nouchg "${index}.lock" 2>/dev/null || true
        rm -f "${index}.lock" 2>/dev/null || true
      fi
    } >>"$logf" 2>&1 || true
  done
}

# Deterministic single-shot repro of the exact message (macOS).
# Useful as a control: proves the error string without needing a race.
repro_exact_once() {
  local repo="$1" logf="$2"
  local gdir index
  gdir="$(abs_git_dir "$repo")"
  index="$gdir/index"
  if [[ "$(uname -s)" != "Darwin" ]]; then
    echo "repro_exact_once: needs Darwin chflags" >>"$logf"
    return 1
  fi
  echo "control $(date +%s%N)" >>"$repo/README.md"
  git_env git -C "$repo" add README.md >>"$logf" 2>&1 || true
  if ! chflags uchg "$index" >>"$logf" 2>&1; then
    echo "chflags uchg failed on $index" >>"$logf"
    return 1
  fi
  {
    git_env git -C "$repo" commit -m 'exact-repro-control'
  } >>"$logf" 2>&1 || true
  chflags nouchg "$index" >>"$logf" 2>&1 || true
  rm -f "${index}.lock" 2>/dev/null || true
  grep -E -q "$EXACT_PATTERN" "$logf"
}

# Kill git mid-flight to interrupt index lock write/rename paths.
worker_kill_mid() {
  local repo="$1" logf="$2" id="$3" n="$4"
  local i f pid
  for (( i = 0; i < n; i++ )); do
    f="files/f-$(printf '%04d' "$(( (RANDOM % FILES) + 1 ))").txt"
    echo "killmid w${id} $i $RANDOM" >>"$repo/$f" 2>/dev/null || true
    git_env git -C "$repo" add -A >>"$logf" 2>&1 || true
    # Start commit in background, kill quickly.
    git_env git -C "$repo" commit -m "killmid-w${id}-$i" >>"$logf" 2>&1 &
    pid=$!
    # Alternate between short sleep and immediate kill to catch different phases.
    if (( i % 3 == 0 )); then
      kill -9 "$pid" 2>/dev/null || true
    elif (( i % 3 == 1 )); then
      sleep 0.001
      kill -9 "$pid" 2>/dev/null || true
    else
      sleep 0.01
      kill -TERM "$pid" 2>/dev/null || true
      sleep 0.005
      kill -9 "$pid" 2>/dev/null || true
    fi
    wait "$pid" 2>/dev/null || true
    # Immediate follow-up write often surfaces "unable to write new index file"
    # if the previous process left a half-written lock/index.
    {
      git_env git -C "$repo" add -A
      git_env git -C "$repo" status -sb
      git_env git -C "$repo" commit -m "after-kill-w${id}-$i" --allow-empty
    } >>"$logf" 2>&1 || true
  done
}

run_workers_for_mode() {
  local mode="$1" round_dir="$2"
  local repo="$ACTIVE_REPO"
  WORKER_PIDS=()
  local w logf

  case "$mode" in
    commit-add)
      for (( w = 0; w < WORKERS; w++ )); do
        logf="$round_dir/w${w}.log"
        if (( w % 2 == 0 )); then
          worker_commit_add "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        else
          worker_add_churn "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        fi
        WORKER_PIDS+=("$!")
      done
      ;;
    commit-commit)
      for (( w = 0; w < WORKERS; w++ )); do
        logf="$round_dir/w${w}.log"
        worker_commit_add "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        WORKER_PIDS+=("$!")
      done
      ;;
    commit-refresh)
      for (( w = 0; w < WORKERS; w++ )); do
        logf="$round_dir/w${w}.log"
        if (( w == 0 )); then
          worker_commit_add "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        else
          worker_refresh "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        fi
        WORKER_PIDS+=("$!")
      done
      ;;
    hook-rewrite)
      for (( w = 0; w < WORKERS; w++ )); do
        logf="$round_dir/w${w}.log"
        if (( w == 0 )); then
          worker_commit_add "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        else
          worker_add_churn "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        fi
        WORKER_PIDS+=("$!")
      done
      ;;
    kill-mid)
      for (( w = 0; w < WORKERS; w++ )); do
        logf="$round_dir/w${w}.log"
        if (( w % 2 == 0 )); then
          worker_kill_mid "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        else
          worker_add_churn "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        fi
        WORKER_PIDS+=("$!")
      done
      ;;
    immutable-index)
      # 1 flapper (long holds) + N-1 commit/add writers
      for (( w = 0; w < WORKERS; w++ )); do
        logf="$round_dir/w${w}.log"
        if (( w == 0 )); then
          worker_immutable_flip "$repo" "$logf" "$w" "$(( OPS_PER_WORKER * 5 ))" &
        elif (( w % 2 == 0 )); then
          worker_commit_add "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        else
          worker_add_churn "$repo" "$logf" "$w" "$OPS_PER_WORKER" &
        fi
        WORKER_PIDS+=("$!")
      done
      ;;
    exact-control)
      # Single deterministic attempt of the exact fatal (no race needed).
      logf="$round_dir/control.log"
      repro_exact_once "$repo" "$logf" || true
      return 0
      ;;
    all)
      for (( w = 0; w < WORKERS; w++ )); do
        logf="$round_dir/w${w}.log"
        case $(( w % 6 )) in
          0) worker_commit_add "$repo" "$logf" "$w" "$OPS_PER_WORKER" & ;;
          1) worker_add_churn "$repo" "$logf" "$w" "$OPS_PER_WORKER" & ;;
          2) worker_refresh "$repo" "$logf" "$w" "$OPS_PER_WORKER" & ;;
          3) worker_commit_only "$repo" "$logf" "$w" "$OPS_PER_WORKER" & ;;
          4) worker_kill_mid "$repo" "$logf" "$w" "$OPS_PER_WORKER" & ;;
          5) worker_immutable_flip "$repo" "$logf" "$w" "$(( OPS_PER_WORKER * 10 ))" & ;;
        esac
        WORKER_PIDS+=("$!")
      done
      ;;
  esac

  local p
  for p in "${WORKER_PIDS[@]}"; do
    wait "$p" 2>/dev/null || true
  done
  WORKER_PIDS=()
}

preserve_hit() {
  local round="$1" round_dir="$2" kind="$3"
  local dest="$HIT_DIR/round-$round-$kind"
  mkdir -p "$dest"
  cp -a "$round_dir"/. "$dest/" 2>/dev/null || true
  {
    echo "reproduced_at_round=$round"
    echo "hit_kind=$kind"
    echo "match_policy=$MATCH_POLICY"
    echo "mode=$MODE"
    echo "workers=$WORKERS"
    echo "files=$FILES"
    echo "linked_worktree=$LINKED_WORKTREE"
    echo "active_repo=$ACTIVE_REPO"
    echo "hit_log=$HIT_LOG"
    echo "time=$(date -Iseconds 2>/dev/null || date)"
    echo "git=$(git --version)"
    echo "--- hit log ---"
    cat "$HIT_LOG" 2>/dev/null || true
    echo "--- matching lines ---"
    grep -E -n -e 'unable to write new index file' -e 'index\.lock' -e 'Unable to write index' \
      "$HIT_LOG" 2>/dev/null || true
    echo "--- git status ---"
    git_env git -C "$ACTIVE_REPO" status 2>&1 || true
    echo "--- git dir ---"
    git_env git -C "$ACTIVE_REPO" rev-parse --git-dir --git-common-dir --show-toplevel 2>&1 || true
  } >"$dest/SUMMARY.txt"
  HIT_SUMMARY="$dest/SUMMARY.txt"
}

# --- main ---

log "git-index race harness"
log "  workdir:  $WORKDIR"
log "  mode:     $MODE"
log "  policy:   $MATCH_POLICY"
log "  max:      $MAX"
log "  workers:  $WORKERS"
log "  files:    $FILES"
log "  ops/w:    $OPS_PER_WORKER"
log "  linked:   $LINKED_WORKTREE"
log "  git:      $(git --version)"
log ""

setup_repo
log "  repo:     $ACTIVE_REPO"
log ""

# Fast path: deterministic exact message (macOS).
if [[ "$MODE" == "exact-control" ]]; then
  round_dir="$ROUNDS_DIR/r-00001"
  mkdir -p "$round_dir"
  logf="$round_dir/control.log"
  if repro_exact_once "$ACTIVE_REPO" "$logf"; then
    HIT_LOG="$logf"
    HIT_KIND="exact"
    preserve_hit 1 "$round_dir" "exact"
    log ""
    log "REPRODUCED (exact) via exact-control"
    log "  hit log:  $HIT_LOG"
    log "  summary:  $HIT_SUMMARY"
    log "  workdir:  $WORKDIR"
    log ""
    log "Matching lines:"
    grep -E -n -e 'unable to write new index file' "$HIT_LOG" | head -5
    exit 0
  fi
  err "exact-control failed to produce the expected message (need macOS chflags?)"
  exit 2
fi

HIT_LOG=""
HIT_SUMMARY=""
HIT_KIND="none"
RELATED_HITS=0
round=0
start_ts=$(date +%s)

while (( round < MAX )); do
  round=$((round + 1))
  round_dir="$ROUNDS_DIR/r-$(printf '%05d' "$round")"
  mkdir -p "$round_dir"

  vlog "round $round: reset"
  reset_round

  vlog "round $round: workers"
  run_workers_for_mode "$MODE" "$round_dir"

  if scan_round_logs "$round_dir"; then
    if [[ "$HIT_KIND" == "exact" ]]; then
      preserve_hit "$round" "$round_dir" "exact"
      elapsed=$(( $(date +%s) - start_ts ))
      log ""
      log "REPRODUCED (exact) at round $round (${elapsed}s)"
      log "  message:  fatal: unable to write new index file"
      log "  hit log:  $HIT_LOG"
      log "  summary:  $HIT_SUMMARY"
      log "  workdir:  $WORKDIR"
      log ""
      log "Matching lines:"
      grep -E -n -e 'unable to write new index file' -e 'index\.lock' \
        "$HIT_LOG" 2>/dev/null | head -20 || true
      if (( ! KEEP_ON_HIT )); then
        rm -rf "$WORKDIR"
        log "  (workdir removed: --no-keep-on-hit)"
      fi
      exit 0
    fi

    # related hit
    RELATED_HITS=$((RELATED_HITS + 1))
    case "$MATCH_POLICY" in
      accept-related)
        preserve_hit "$round" "$round_dir" "related"
        elapsed=$(( $(date +%s) - start_ts ))
        log ""
        log "REPRODUCED (related) at round $round (${elapsed}s)"
        log "  kind:     index.lock contention (same failure family)"
        log "  hint:     use --strict or --until-exact for the exact message"
        log "  hit log:  $HIT_LOG"
        log "  summary:  $HIT_SUMMARY"
        log "  workdir:  $WORKDIR"
        log ""
        log "Matching lines:"
        grep -E -n -e 'unable to write new index file' -e 'index\.lock' \
          "$HIT_LOG" 2>/dev/null | head -20 || true
        if (( ! KEEP_ON_HIT )); then
          rm -rf "$WORKDIR"
        fi
        exit 0
        ;;
      until-exact)
        vlog "round $round: related hit (#$RELATED_HITS), continuing for exact"
        # keep a copy of first related for reference
        if (( RELATED_HITS == 1 )); then
          preserve_hit "$round" "$round_dir" "related-first"
        fi
        ;;
      strict)
        vlog "round $round: related hit ignored (--strict)"
        ;;
    esac
  fi

  if (( round > 5 )); then
    rm -rf "$ROUNDS_DIR/r-$(printf '%05d' "$((round - 5))")" 2>/dev/null || true
  fi

  if (( round % 50 == 0 )) || (( VERBOSE )); then
    elapsed=$(( $(date +%s) - start_ts ))
    log "  … round $round / $MAX (${elapsed}s, related_hits=$RELATED_HITS, no exact yet)"
  fi
done

elapsed=$(( $(date +%s) - start_ts ))
log ""
if (( RELATED_HITS > 0 )); then
  log "NOT REPRODUCED exact message in $MAX rounds (${elapsed}s)"
  log "  related index.lock hits seen: $RELATED_HITS"
  log "  (contention is real; exact write failure may need kill-mid / FS pressure)"
else
  log "NOT REPRODUCED in $MAX rounds (${elapsed}s)"
fi
log "  workdir left at: $WORKDIR"
exit 1

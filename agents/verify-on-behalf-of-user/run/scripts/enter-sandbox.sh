#!/usr/bin/env bash
# enter-sandbox.sh — set up ~/.sandbox for verify-on-behalf-of-user
# Usage: source .agents/skills/verify-on-behalf-of-user/scripts/enter-sandbox.sh [--no-reset]
set -euo pipefail

# Resolve the real user home even if HOME was already pointed at the sandbox.
if [[ -n "${VERIFY_REAL_HOME:-}" ]]; then
  _REAL_HOME="${VERIFY_REAL_HOME}"
elif [[ "${HOME:-}" == *"/.sandbox/default-home"* ]]; then
  _REAL_HOME="${HOME%%/.sandbox/default-home*}"
elif [[ -n "${HOME:-}" ]]; then
  _REAL_HOME="${HOME}"
else
  _REAL_HOME="$(eval echo "~${USER}")"
fi
SANDBOX_ROOT="${_REAL_HOME}/.sandbox"
SANDBOX_HOME="${SANDBOX_ROOT}/default-home"
SANDBOX_BIN="${SANDBOX_ROOT}/bin"
SANDBOX_TRANSCRIPTS="${SANDBOX_ROOT}/transcripts"

RESET=1
for arg in "$@"; do
  case "$arg" in
    --no-reset) RESET=0 ;;
    -h|--help)
      echo "Usage: source enter-sandbox.sh [--no-reset]"
      echo "  Sets HOME=${SANDBOX_HOME}, PATH=${SANDBOX_BIN}:..."
      echo "  Default: reset sandbox data (.tsk, .config) under default-home"
      return 0 2>/dev/null || exit 0
      ;;
  esac
done

mkdir -p "${SANDBOX_HOME}" "${SANDBOX_BIN}" "${SANDBOX_TRANSCRIPTS}"
mkdir -p "${SANDBOX_HOME}/tmp"

if [[ "${RESET}" -eq 1 ]]; then
  rm -rf "${SANDBOX_HOME}/.tsk" "${SANDBOX_HOME}/.config"
fi

export HOME="${SANDBOX_HOME}"
export TMPDIR="${SANDBOX_HOME}/tmp"
export XDG_CONFIG_HOME="${HOME}/.config"
export TSK_HOME="${HOME}/.tsk"
export SANDBOX_ROOT SANDBOX_HOME SANDBOX_BIN SANDBOX_TRANSCRIPTS
export PATH="${SANDBOX_BIN}:${PATH}"

# Keep Go caches under sandbox (optional; avoids polluting user cache during verify)
export GOCACHE="${SANDBOX_ROOT}/gocache"
mkdir -p "${GOCACHE}"

echo "sandbox: HOME=${HOME}"
echo "sandbox: SANDBOX_BIN=${SANDBOX_BIN}"
echo "sandbox: transcripts=${SANDBOX_TRANSCRIPTS}"
if [[ "${RESET}" -eq 1 ]]; then
  echo "sandbox: data reset (use --no-reset to keep prior .tsk/.config)"
fi
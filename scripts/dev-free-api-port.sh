#!/usr/bin/env bash
# Free APP_PORT for local dev by stopping only this repo's API (go run ./cmd/api
# or bin/server built here). Linux: systemd APP_SERVICE_UNIT@* (default nextpresskit-backend@*) is detected and
# not killed. macOS: uses lsof/ps (no /proc). Git Bash on Windows: best-effort;
# use scripts/nextpresskit.ps1 run for native Windows.
#
# Usage:
#   dev-free-api-port.sh <port> <repo_root>        # strict: exit 1 if blocked
#   dev-free-api-port.sh <port> <repo_root> soft # for stop: warn, exit 0

set -euo pipefail

port="${1:?usage: dev-free-api-port.sh <port> <repo_root> [soft]}"
workdir="${2:?usage: dev-free-api-port.sh <port> <repo_root> [soft]}"
soft="${3:-}"
workdir="$(cd "$workdir" && pwd -P 2>/dev/null || pwd)"
binserver="$workdir/bin/server"
kernel="$(uname -s 2>/dev/null || echo unknown)"
service_unit="$(awk -F= '/^APP_SERVICE_UNIT=/{print $2; exit}' "$workdir/.env" 2>/dev/null | tr -d '[:space:]')"
[[ -z "$service_unit" ]] && service_unit="nextpresskit-backend"

fail_or_soft() {
  echo "$1" >&2
  if [[ "$soft" == "soft" ]]; then
    exit 0
  fi
  exit 1
}

# --- Linux (ss + /proc) -----------------------------------------------------

port_busy_linux() {
  ss -ltn "( sport = :$port )" 2>/dev/null | awk 'NR>1 { found=1 } END { exit found ? 0 : 1 }'
}

listener_pids_linux() {
  # ss often omits pid= without root; lsof usually still lists the listener for your own processes.
  local from_ss from_lsof
  from_ss="$(ss -ltn "( sport = :$port )" 2>/dev/null | sed -n 's/.*pid=\([0-9][0-9]*\).*/\1/p' | sort -u)"
  from_lsof=""
  if command -v lsof >/dev/null 2>&1; then
    from_lsof="$(lsof -nP -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null | sort -u)"
  fi
  printf '%s\n%s\n' "$from_ss" "$from_lsof" | awk 'NF' | sort -u
}

is_nextpresskit_dev_api_linux() {
  local pid="$1"
  [[ "$pid" =~ ^[0-9]+$ ]] || return 1
  [[ -r "/proc/$pid/cmdline" ]] || return 1
  local exe=""
  exe="$(readlink -f "/proc/$pid/exe" 2>/dev/null || readlink "/proc/$pid/exe" 2>/dev/null || true)"
  # Rebuilt binary while process runs: kernel appends " (deleted)" to /proc/PID/exe text.
  exe="${exe% (deleted)}"
  if [[ -n "$exe" && "$exe" == "$binserver" ]]; then
    return 0
  fi
  local cl=""
  cl="$(tr '\0' ' ' <"/proc/$pid/cmdline" 2>/dev/null || true)"
  [[ "$cl" == *"$workdir/bin/server"* ]] && return 0
  [[ "$cl" == *"$workdir/cmd/api"* ]] && return 0
  [[ "$cl" == *./cmd/api* ]] && [[ "$cl" == *go\ run* ]] && return 0
  return 1
}

# --- macOS / BSD (lsof + ps) ------------------------------------------------

port_busy_darwin() {
  command -v lsof >/dev/null 2>&1 || return 1
  lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1
}

listener_pids_darwin() {
  command -v lsof >/dev/null 2>&1 || return 0
  lsof -nP -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null | sort -u
}

is_nextpresskit_dev_api_darwin() {
  local pid="$1"
  [[ "$pid" =~ ^[0-9]+$ ]] || return 1
  local args=""
  args="$(ps -p "$pid" -o args= 2>/dev/null || true)"
  [[ -z "${args// }" ]] && return 1
  [[ "$args" == *"$workdir/cmd/api"* ]] && return 0
  [[ "$args" == *./cmd/api* ]] && [[ "$args" == *go* ]] && [[ "$args" == *run* ]] && return 0
  if [[ "$args" == *"$workdir/bin/server"* ]] || [[ "$args" == *"$workdir/bin/server "* ]]; then
    return 0
  fi
  # argv0 might be path to server
  [[ "$(basename "${args%% *}")" == "server" ]] && [[ "$args" == *"$workdir"* ]] && return 0
  return 1
}

# --- dispatch ----------------------------------------------------------------

port_busy() {
  case "$kernel" in
  Linux*) port_busy_linux ;;
  Darwin*) port_busy_darwin ;;
  *)
    if command -v ss >/dev/null 2>&1; then
      port_busy_linux
    elif command -v lsof >/dev/null 2>&1; then
      port_busy_darwin
    else
      return 1
    fi
    ;;
  esac
}

listener_pids() {
  case "$kernel" in
  Linux*) listener_pids_linux ;;
  Darwin*) listener_pids_darwin ;;
  *)
    if command -v ss >/dev/null 2>&1; then
      listener_pids_linux
    elif command -v lsof >/dev/null 2>&1; then
      listener_pids_darwin
    else
      echo ""
    fi
    ;;
  esac
}

is_nextpresskit_dev_api() {
  case "$kernel" in
  Linux*) is_nextpresskit_dev_api_linux "$1" ;;
  Darwin*) is_nextpresskit_dev_api_darwin "$1" ;;
  *)
    if [[ -r "/proc/$1/cmdline" ]]; then
      is_nextpresskit_dev_api_linux "$1"
    else
      is_nextpresskit_dev_api_darwin "$1"
    fi
    ;;
  esac
}

systemd_unit_for_pid() {
  local pid="$1"
  command -v systemctl >/dev/null 2>&1 || return 1
  systemctl show --pid="$pid" -p Id --value 2>/dev/null | awk 'NF && $0 != "-" { print; exit 0 }'
  return 1
}

is_repo_systemd_unit() {
  local u="$1"
  case "$u" in
    "${service_unit}"@*.service) return 0 ;;
    */"${service_unit}"@*.service) return 0 ;;
    *) return 1 ;;
  esac
}

pids="$(listener_pids || true)"
pids="${pids//$'\n'/ }"

if [[ -z "${pids// }" ]]; then
  if port_busy 2>/dev/null; then
    fail_or_soft "Port $port is in use but listener PID is hidden (try: sudo lsof -nP -iTCP:$port -sTCP:LISTEN or sudo ss -ltnp '( sport = :$port )')."
  fi
  exit 0
fi

our=""
foreign=""
for pid in $pids; do
  [[ -z "${pid// }" ]] && continue
  unit="$(systemd_unit_for_pid "$pid" || true)"
  if [[ -n "$unit" ]] && is_repo_systemd_unit "$unit" && is_nextpresskit_dev_api "$pid"; then
    echo "Port $port: API is managed by systemd ($unit), pid=$pid." >&2
    echo "Killing it would only respawn the service. Run:" >&2
    echo "  sudo systemctl stop $unit" >&2
    if [[ "$soft" == "soft" ]]; then
      exit 0
    fi
    exit 1
  fi
  if is_nextpresskit_dev_api "$pid"; then
    our="$our $pid"
  else
    foreign="$foreign $pid"
  fi
done

if [[ -n "${foreign// }" ]]; then
  echo "Port $port is held by another process:" >&2
  for pid in $foreign; do
    ps -p "$pid" -o pid,user,args 2>/dev/null || true
  done
  fail_or_soft "Change APP_PORT in .env or stop that process."
fi

if [[ -z "${our// }" ]]; then
  if port_busy 2>/dev/null; then
    fail_or_soft "Port $port is in use; could not match listener to this repo's dev API."
  fi
  exit 0
fi

for pid in $our; do
  echo "Stopping dev API on port $port (pid=$pid)..." >&2
  kill -TERM "$pid" 2>/dev/null || true
done

for _ in $(seq 1 15); do
  if ! port_busy 2>/dev/null; then
    exit 0
  fi
  sleep 0.2
done

for pid in $our; do
  if kill -0 "$pid" 2>/dev/null; then
    echo "Process $pid did not exit; sending SIGKILL." >&2
    kill -KILL "$pid" 2>/dev/null || true
  fi
done

sleep 0.3
if port_busy 2>/dev/null; then
  fail_or_soft "Port $port is still in use after stop attempt."
fi
exit 0

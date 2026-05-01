#!/usr/bin/env bash
# Background go run API with pid file + logs
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export NP_ROOT="$ROOT"
# shellcheck source=scripts/lib/app-port.sh
source "$ROOT/scripts/lib/app-port.sh"

_dev_rt="$(np_dev_runtime_basename)"
API_PID_FILE="${API_PID_FILE:-$ROOT/.tmp/${_dev_rt}-api.pid}"
API_LOG_FILE="${API_LOG_FILE:-$ROOT/.tmp/${_dev_rt}-api.log}"
mkdir -p "$(dirname "$API_PID_FILE")"

if [[ -f "$API_PID_FILE" ]]; then
  pid="$(cat "$API_PID_FILE")"
  if kill -0 "$pid" 2>/dev/null; then
    echo "API already running (pid=$pid)."
    exit 0
  fi
  rm -f "$API_PID_FILE"
fi

port="$(np_app_port)"
bash "$ROOT/scripts/dev-free-api-port.sh" "$port" "$ROOT" || exit 1

port_still_busy() {
  if command -v ss >/dev/null 2>&1 && ss -ltn "( sport = :$port )" 2>/dev/null | awk 'NR>1 { found=1 } END { exit found ? 0 : 1 }'; then
    return 0
  fi
  if command -v lsof >/dev/null 2>&1 && lsof -nP -tiTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
    return 0
  fi
  return 1
}

if port_still_busy; then
  echo "Port $port is still in use after attempting to free it." >&2
  exit 1
fi

nohup go run ./cmd/api >"$API_LOG_FILE" 2>&1 &
echo $! >"$API_PID_FILE"
sleep 1
pid="$(cat "$API_PID_FILE")"
if kill -0 "$pid" 2>/dev/null; then
  echo "API started in background (pid=$pid)."
  echo "Logs: $API_LOG_FILE"
else
  echo "API failed to start. See $API_LOG_FILE" >&2
  rm -f "$API_PID_FILE"
  exit 1
fi

#!/usr/bin/env bash
# Stop make-start API and soft-clear stray same-repo listeners on APP_PORT
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export NP_ROOT="$ROOT"
# shellcheck source=scripts/lib/app-port.sh
source "$ROOT/scripts/lib/app-port.sh"

_dev_rt="$(np_dev_runtime_basename)"
API_PID_FILE="${API_PID_FILE:-$ROOT/.tmp/${_dev_rt}-api.pid}"

if [[ -f "$API_PID_FILE" ]]; then
  pid="$(cat "$API_PID_FILE")"
  if kill -0 "$pid" 2>/dev/null; then
    kill -TERM "$pid"
    stopped=0
    for _ in 1 2 3 4 5 6 7 8 9 10; do
      if ! kill -0 "$pid" 2>/dev/null; then
        rm -f "$API_PID_FILE"
        echo "API stopped."
        stopped=1
        break
      fi
      sleep 1
    done
    if [[ "$stopped" != 1 ]]; then
      echo "API did not stop in time; sending SIGKILL."
      kill -KILL "$pid" 2>/dev/null || true
      rm -f "$API_PID_FILE"
      echo "API stopped."
    fi
  else
    echo "API is not running (stale pid file)."
    rm -f "$API_PID_FILE"
  fi
else
  echo "No make-start pid file ($API_PID_FILE)."
fi

port="$(np_app_port)"
bash "$ROOT/scripts/dev-free-api-port.sh" "$port" "$ROOT" soft || true

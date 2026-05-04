#!/usr/bin/env bash
# Foreground API: free APP_PORT for this repo, then go run ./cmd/api
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
export NP_ROOT="$ROOT"
# shellcheck source=scripts/lib/app-port.sh
source "$ROOT/scripts/lib/app-port.sh"
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

interrupted=0
trap 'interrupted=1' INT
go run ./cmd/api
status=$?
if [[ $interrupted -eq 1 || $status -eq 130 ]]; then
  exit 0
fi
if [[ $status -ne 0 ]]; then
  echo "run finished with exit code $status" >&2
fi
exit 0

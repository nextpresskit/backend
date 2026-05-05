#!/usr/bin/env bash
# Text-only setup menu for make setup / nextpresskit setup (no gum/dialog).
# NP_SETUP_NONINTERACTIVE=1 or non-TTY: linear install → build-all → migrate-up → seed (+ optional local HTTPS).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
NP="$ROOT/scripts/nextpresskit"
cd "$ROOT"

STEP_ORDER=(
  install deps tidy clean build build-all
  migrate-drop db-fresh migrate-up seed
  test test-coverage test-integration security-check checks
  local-https deploy-nginx deploy start stop
)

# Registry order (see internal/appregistry + docs/MODULES.md)
VALID_MODULES=(user rbac auth taxonomy media posts pages)

read_modules_value() {
  [[ -f "$ROOT/.env" ]] || { echo ""; return 0; }
  awk '/^[[:space:]]*MODULES[[:space:]]*=/ {
    line=$0
    sub(/^[[:space:]]*MODULES[[:space:]]*=/, "", line)
    gsub(/\r/, "", line)
    gsub(/^[[:space:]]+|[[:space:]]+$/, "", line)
    print line
    exit
  }' "$ROOT/.env"
}

set_modules_value() {
  local newval=$1
  local envf="$ROOT/.env"
  if [[ ! -f "$envf" ]]; then
    echo "No .env — run option 1 (full setup) or: ./scripts/nextpresskit install" >&2
    return 1
  fi
  awk -v nv="$newval" '
    /^[[:space:]]*MODULES[[:space:]]*=/ { print "MODULES=" nv; rep=1; next }
    { print }
    END { if (!rep) print "MODULES=" nv }
  ' "$envf" > "${envf}.np.tmp" && mv "${envf}.np.tmp" "$envf"
}

token_in_list() {
  local needle=$1
  shift
  local h hl
  needle=$(echo "$needle" | tr '[:upper:]' '[:lower:]')
  for h in "$@"; do
    hl=$(echo "$h" | tr '[:upper:]' '[:lower:]' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    [[ "$hl" == "$needle" ]] && return 0
  done
  return 1
}

run_modules_add() {
  local raw pick idx m i
  raw="$(read_modules_value)"
  raw="${raw//$'\r'/}"
  raw="${raw#"${raw%%[![:space:]]*}"}"
  raw="${raw%"${raw##*[![:space:]]}"}"
  if [[ -z "${raw// }" ]]; then
    echo "MODULES empty = all modules. Use option 7 first for an explicit list, then 6 to add." >&2
    return 0
  fi
  local IFS=,
  read -ra have <<< "$(echo "$raw" | tr '[:upper:]' '[:lower:]')"
  local cleaned=()
  for m in "${have[@]}"; do
    m=$(echo "$m" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    [[ -n "$m" ]] && cleaned+=("$m")
  done
  local avail=()
  for m in "${VALID_MODULES[@]}"; do
    token_in_list "$m" "${cleaned[@]}" && continue
    avail+=("$m")
  done
  if [[ ${#avail[@]} -eq 0 ]]; then
    echo "All ids already in MODULES." >&2
    return 0
  fi
  echo "Add MODULES id:" >&2
  for i in "${!avail[@]}"; do
    printf " %d) %s\n" "$((i + 1))" "${avail[i]}" >&2
  done
  read -r -p "Number [0=cancel]: " pick || true
  [[ "${pick:-0}" == "0" || -z "${pick// }" ]] && return 0
  [[ "$pick" =~ ^[0-9]+$ ]] || { echo "Invalid." >&2; return 1; }
  idx=$((pick - 1))
  [[ "$idx" -ge 0 && "$idx" -lt ${#avail[@]} ]] || { echo "Invalid." >&2; return 1; }
  local add="${avail[idx]}"
  local newcsv="${raw},${add}"
  set_modules_value "$newcsv" || return 1
  echo "MODULES=$newcsv" >&2
  echo "Then: migrate-up, seed. docs/MODULES.md" >&2
}

run_modules_remove() {
  local raw pick idx m i rem j
  raw="$(read_modules_value)"
  raw="${raw//$'\r'/}"
  raw="${raw#"${raw%%[![:space:]]*}"}"
  raw="${raw%"${raw##*[![:space:]]}"}"
  local newcsv=""
  if [[ -z "${raw// }" ]]; then
    echo "MODULES empty = full kit. Remove one (rest → .env):" >&2
    for i in "${!VALID_MODULES[@]}"; do
      printf " %d) %s\n" "$((i + 1))" "${VALID_MODULES[i]}" >&2
    done
    read -r -p "Number [0=cancel]: " pick || true
    [[ "${pick:-0}" == "0" || -z "${pick// }" ]] && return 0
    [[ "$pick" =~ ^[0-9]+$ ]] || { echo "Invalid." >&2; return 1; }
    idx=$((pick - 1))
    [[ "$idx" -ge 0 && "$idx" -lt ${#VALID_MODULES[@]} ]] || { echo "Invalid." >&2; return 1; }
    rem="${VALID_MODULES[idx]}"
    local parts=()
    for m in "${VALID_MODULES[@]}"; do
      [[ "$m" == "$rem" ]] && continue
      parts+=("$m")
    done
    local IFS=,
    newcsv="${parts[*]}"
  else
    local IFS=,
    read -ra cleaned <<< "$(echo "$raw" | tr '[:upper:]' '[:lower:]')"
    local toks=()
    for m in "${cleaned[@]}"; do
      m=$(echo "$m" | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
      [[ -n "$m" ]] && toks+=("$m")
    done
    if [[ ${#toks[@]} -eq 0 ]]; then
      echo "Could not parse MODULES." >&2
      return 1
    fi
    echo "MODULES tokens — remove one:" >&2
    for j in "${!toks[@]}"; do
      printf " %d) %s\n" "$((j + 1))" "${toks[j]}" >&2
    done
    read -r -p "Number [0=cancel]: " pick || true
    [[ "${pick:-0}" == "0" || -z "${pick// }" ]] && return 0
    [[ "$pick" =~ ^[0-9]+$ ]] || { echo "Invalid." >&2; return 1; }
    idx=$((pick - 1))
    [[ "$idx" -ge 0 && "$idx" -lt ${#toks[@]} ]] || { echo "Invalid." >&2; return 1; }
    local out=()
    for j in "${!toks[@]}"; do
      [[ "$j" -eq "$idx" ]] && continue
      out+=("${toks[j]}")
    done
    IFS=,
    newcsv="${out[*]}"
  fi
  set_modules_value "$newcsv" || return 1
  if [[ -z "${newcsv// }" ]]; then
    echo "MODULES= (empty → full default)" >&2
  else
    echo "MODULES=$newcsv" >&2
  fi
  echo "migrate-up if needed. docs/MODULES.md" >&2
}

step_hint() {
  case "$1" in
  install) echo "go mod download; .env from .env.example if missing" ;;
  deps) echo "go mod download only" ;;
  tidy) echo "go mod tidy" ;;
  clean) echo "rm bin/; go clean" ;;
  build) echo "go build → bin/server" ;;
  build-all) echo "bin/server, bin/migrate, bin/seed" ;;
  migrate-up) echo "cmd/migrate up (MODULES)" ;;
  seed) echo "cmd/seed" ;;
  migrate-drop) echo "drop public tables (confirm)" ;;
  db-fresh) echo "migrate-drop + migrate-up" ;;
  test) echo "go test -v ./..." ;;
  test-coverage) echo "go test -cover ./..." ;;
  test-integration) echo "go test -tags=integration (needs DB_*)" ;;
  security-check) echo "govulncheck ./..." ;;
  checks) echo "test, vet, integration, openapi validate, govulncheck" ;;
  local-https) echo "scripts/setup-local-https.sh" ;;
  deploy-nginx) echo "scripts/deploy apply-nginx" ;;
  deploy) echo "scripts/deploy" ;;
  start) echo "background API (Unix)" ;;
  stop) echo "stop background API" ;;
  *) echo "$1" ;;
  esac
}

run_linear_setup() {
  "$NP" install
  "$NP" build-all
  "$NP" migrate-up
  "$NP" seed
  local uname_s
  uname_s="$(uname -s 2>/dev/null || echo unknown)"
  case "$uname_s" in
  MINGW* | MSYS* | CYGWIN*) ;;
  *)
    if [[ -t 0 ]] && [[ "${SKIP_SETUP_LOCAL_HTTPS:-}" != "1" ]]; then
      bash "$ROOT/scripts/setup-local-https.sh" || true
    fi
    ;;
  esac
  echo "Setup complete. Run: ./scripts/nextpresskit run   (or: make run)" >&2
}

is_destructive() {
  case "$1" in
  db-fresh | migrate-drop) return 0 ;;
  *) return 1 ;;
  esac
}

run_one_step() {
  local id=$1
  case "$id" in
  local-https) bash "$ROOT/scripts/setup-local-https.sh" || true ;;
  deploy) bash "$ROOT/scripts/deploy" ;;
  deploy-nginx) bash "$ROOT/scripts/deploy" apply-nginx ;;
  *) "$NP" "$id" ;;
  esac
}

confirm_yes() {
  local msg=$1
  local ans
  if [[ -r /dev/tty ]]; then
    read -r -p "$msg [y/N]: " ans </dev/tty || return 1
  else
    read -r -p "$msg [y/N]: " ans || return 1
  fi
  [[ "${ans:-}" == "y" || "${ans:-}" == "Y" ]]
}

order_selected_steps() {
  local sel=$1
  local db_fresh=0
  echo "$sel" | grep -qxF db-fresh && db_fresh=1 || true
  local id
  for id in "${STEP_ORDER[@]}"; do
    echo "$sel" | grep -qxF "$id" || continue
    if [[ "$db_fresh" == 1 ]]; then
      [[ "$id" == migrate-drop ]] && continue
      [[ "$id" == migrate-up ]] && continue
    fi
    echo "$id"
  done
}

execute_plan() {
  local plan=$1
  local id hint
  while IFS= read -r id; do
    [[ -z "$id" ]] && continue
    if is_destructive "$id"; then
      hint="$(step_hint "$id")"
      confirm_yes "Run '$id' — $hint?" || { echo "(skipped $id)" >&2; continue; }
    fi
    echo "==> $id" >&2
    run_one_step "$id"
  done <<<"$plan"
}

profile_steps() {
  case "$1" in
  1 | full) printf '%s\n' install build-all migrate-up seed local-https ;;
  2 | db) printf '%s\n' migrate-up seed ;;
  3 | build) printf '%s\n' install build-all ;;
  4 | quality) printf '%s\n' checks ;;
  *) return 1 ;;
  esac
}

prompt_profile() {
  # Menus on stderr so they show immediately when stdout is block-buffered (e.g. under make).
  echo "" >&2
  echo "NextPressKit setup" >&2
  echo "------------------" >&2
  echo " 1) Full setup — install, build-all, migrate-up, seed; mkcert/HTTPS if interactive (skip when piped/CI)" >&2
  echo " 2) Refresh database — migrate-up + seed (after pulling code or changing MODULES)" >&2
  echo " 3) Compile only — install + build-all (binaries; no DB)" >&2
  echo " 4) Quality gate — tests, vet, integration, OpenAPI check, govulncheck" >&2
  echo " 5) Custom — run chosen nextpresskit steps (numbered list)" >&2
  echo " 6) Enable a kit module — add id to MODULES in .env (then run migrate-up / seed)" >&2
  echo " 7) Disable a kit module — remove id from MODULES in .env (then migrate-up if needed)" >&2
  echo " 8) Exit" >&2
  echo "" >&2
  local choice
  read -r -p "[1-8]: " choice || true
  echo "$choice"
}

prompt_custom() {
  echo "" >&2
  echo "Step numbers (space/comma). 0 or empty = back." >&2
  local i=0
  local id
  for id in "${STEP_ORDER[@]}"; do
    i=$((i + 1))
    printf " %2d) %-14s %s\n" "$i" "$id" "$(step_hint "$id")" >&2
  done
  echo "" >&2
  local line
  read -r -p "> " line || true
  [[ -z "${line// }" ]] && return 0
  if [[ "$line" =~ ^[[:space:]]*0[[:space:]]*$ ]]; then
    return 0
  fi
  local chosen="" num idx
  line="${line//,/ }"
  for num in $line; do
    [[ "$num" =~ ^[0-9]+$ ]] || continue
    idx=$((num - 1))
    if [[ "$idx" -ge 0 && "$idx" -lt ${#STEP_ORDER[@]} ]]; then
      chosen+="${STEP_ORDER[$idx]}"$'\n'
    fi
  done
  if [[ -z "$chosen" ]]; then
    echo "No valid step numbers (use 1-${#STEP_ORDER[@]})." >&2
    return 0
  fi
  echo -n "$chosen"
}

interactive_main() {
  local choice steps ordered
  while true; do
    choice="$(prompt_profile)"
    case "${choice:-}" in
    1 | 2 | 3 | 4)
      steps="$(profile_steps "$choice")"
      ordered="$(order_selected_steps "$steps")"
      ;;
    5)
      steps="$(prompt_custom)"
      if [[ -z "$(echo "$steps" | sed '/^$/d')" ]]; then
        continue
      fi
      ordered="$(order_selected_steps "$steps")"
      if [[ -z "$(echo "$ordered" | sed '/^$/d')" ]]; then
        continue
      fi
      ;;
    6)
      run_modules_add
      exit 0
      ;;
    7)
      run_modules_remove
      exit 0
      ;;
    8 | "") exit 0 ;;
    *)
      echo "Invalid choice (1-8)." >&2
      continue
      ;;
    esac

    if [[ -z "$(echo "$ordered" | sed '/^$/d')" ]]; then
      echo "No steps." >&2
      exit 0
    fi

    execute_plan "$ordered"
    echo "" >&2
    echo "Done. make run | ./scripts/nextpresskit run" >&2
    break
  done
}

if [[ "${NP_SETUP_NONINTERACTIVE:-}" == "1" ]] || ! [[ -t 0 ]]; then
  run_linear_setup
else
  interactive_main
fi

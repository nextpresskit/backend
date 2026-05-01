#!/usr/bin/env bash
# shellcheck shell=bash
# Resolve APP_PORT from repo .env (default 9090). Expects NP_ROOT set to repo root.

np_app_port() {
  local root="${NP_ROOT:-.}"
  local p
  p="$(awk -F= '/^APP_PORT=/{print $2; exit}' "$root/.env" 2>/dev/null | tr -d '[:space:]')"
  if [[ -z "$p" ]]; then
    echo "9090"
  else
    echo "$p"
  fi
}

np_dev_runtime_basename() {
  local root="${NP_ROOT:-.}" b
  b="$(awk -F= '/^APP_DEV_RUNTIME_BASENAME=/{print $2; exit}' "$root/.env" 2>/dev/null | tr -d '[:space:]')"
  [[ -z "$b" ]] && b="nextpresskit"
  echo "$b"
}

np_app_service_unit() {
  local root="${NP_ROOT:-.}" u
  u="$(awk -F= '/^APP_SERVICE_UNIT=/{print $2; exit}' "$root/.env" 2>/dev/null | tr -d '[:space:]')"
  [[ -z "$u" ]] && u="nextpresskit-backend"
  echo "$u"
}

np_local_ssl_subdir() {
  local root="${NP_ROOT:-.}" s
  s="$(awk -F= '/^APP_LOCAL_SSL_SUBDIR=/{print $2; exit}' "$root/.env" 2>/dev/null | tr -d '[:space:]')"
  [[ -z "$s" ]] && s="nextpresskit-ssl"
  echo "$s"
}

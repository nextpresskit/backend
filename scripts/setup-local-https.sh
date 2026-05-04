#!/usr/bin/env bash
# Best-effort local HTTPS: install mkcert if missing, SANs that match browser URLs + Nginx (Linux).
# Invoked from ./scripts/nextpresskit setup / make setup when stdin is a TTY.
# Skip: SKIP_SETUP_LOCAL_HTTPS=1  or  non-interactive (no TTY) setup.
set -u

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ -f "$ROOT/.env" ]]; then
  set -a
  # shellcheck source=/dev/null
  source "$ROOT/.env"
  set +a
fi
LOCAL_HOST="${NEXTPRESS_PUBLIC_HOST:-nextpresskit.local}"

cyan()  { printf '\033[36m%s\033[0m\n' "$*" >&2; }
green() { printf '\033[32m%s\033[0m\n' "$*" >&2; }
yellow(){ printf '\033[33m%s\033[0m\n' "$*" >&2; }
red()   { printf '\033[31m%s\033[0m\n' "$*" >&2; }

SSL_DIR="${XDG_DATA_HOME:-"$HOME/.local/share"}/${APP_LOCAL_SSL_SUBDIR:-nextpresskit-ssl}"
CERT="$SSL_DIR/cert.pem"
KEY="$SSL_DIR/key.pem"

mkdir -p "$SSL_DIR"

# Load Homebrew into this shell when present (so mkcert appears after brew install).
np_brew_shellenv() {
  local p
  if command -v brew >/dev/null 2>&1; then
    eval "$(brew shellenv 2>/dev/null)" || true
    return 0
  fi
  for p in /opt/homebrew/bin/brew /usr/local/bin/brew /home/linuxbrew/.linuxbrew/bin/brew; do
    if [[ -x "$p" ]]; then
      eval "$("$p" shellenv 2>/dev/null)" || true
      return 0
    fi
  done
  return 1
}

# Print path to mkcert binary, or empty.
np_mkcert_path() {
  local p
  p="$(command -v mkcert 2>/dev/null || true)"
  if [[ -n "$p" ]]; then
    printf '%s\n' "$p"
    return 0
  fi
  np_brew_shellenv || true
  p="$(command -v mkcert 2>/dev/null || true)"
  if [[ -n "$p" ]]; then
    printf '%s\n' "$p"
    return 0
  fi
  if command -v brew >/dev/null 2>&1; then
    p="$(brew --prefix 2>/dev/null)/bin/mkcert"
    if [[ -x "$p" ]]; then
      printf '%s\n' "$p"
      return 0
    fi
  fi
  for p in /opt/homebrew/bin/mkcert /usr/local/bin/mkcert /home/linuxbrew/.linuxbrew/bin/mkcert; do
    if [[ -x "$p" ]]; then
      printf '%s\n' "$p"
      return 0
    fi
  done
  return 1
}

# Install mkcert using a known package manager or Homebrew. Returns 0 if mkcert is available after.
ensure_mkcert_installed() {
  if np_mkcert_path >/dev/null; then
    return 0
  fi

  local os
  os="$(uname -s 2>/dev/null || echo "")"

  if [[ "$os" == Linux ]]; then
    if command -v apt-get >/dev/null 2>&1; then
      cyan "mkcert not found; installing via apt-get (sudo)…" >&2
      if sudo DEBIAN_FRONTEND=noninteractive apt-get update -qq &&
        sudo DEBIAN_FRONTEND=noninteractive apt-get install -y mkcert; then
        hash -r 2>/dev/null || true
        np_mkcert_path >/dev/null && return 0
      fi
      yellow "apt install mkcert failed or package missing — try another method below." >&2
    fi
    if command -v dnf >/dev/null 2>&1; then
      cyan "mkcert not found; installing via dnf (sudo)…" >&2
      if sudo dnf install -y mkcert; then
        hash -r 2>/dev/null || true
        np_mkcert_path >/dev/null && return 0
      fi
    fi
    if command -v pacman >/dev/null 2>&1; then
      cyan "mkcert not found; installing via pacman (sudo)…" >&2
      if sudo pacman -S --noconfirm mkcert; then
        hash -r 2>/dev/null || true
        np_mkcert_path >/dev/null && return 0
      fi
    fi
    if command -v zypper >/dev/null 2>&1; then
      cyan "mkcert not found; installing via zypper (sudo)…" >&2
      if sudo zypper --non-interactive install mkcert; then
        hash -r 2>/dev/null || true
        np_mkcert_path >/dev/null && return 0
      fi
    fi
  fi

  local brew_bin=""
  if command -v brew >/dev/null 2>&1; then
    brew_bin="$(command -v brew)"
  elif [[ -x /opt/homebrew/bin/brew ]]; then
    brew_bin="/opt/homebrew/bin/brew"
  elif [[ -x /usr/local/bin/brew ]]; then
    brew_bin="/usr/local/bin/brew"
  elif [[ -x /home/linuxbrew/.linuxbrew/bin/brew ]]; then
    brew_bin="/home/linuxbrew/.linuxbrew/bin/brew"
  fi

  if [[ -n "$brew_bin" ]]; then
    cyan "mkcert not found; installing via Homebrew…" >&2
    if "$brew_bin" install mkcert; then
      eval "$("$brew_bin" shellenv 2>/dev/null)" || true
      hash -r 2>/dev/null || true
      np_mkcert_path >/dev/null && return 0
    fi
  fi

  return 1
}

cyan "=== Local HTTPS (optional) ===" >&2

MKCERT=""
if ! MKCERT="$(np_mkcert_path)"; then
  if ! ensure_mkcert_installed; then
    MKCERT=""
  else
    MKCERT="$(np_mkcert_path)" || MKCERT=""
  fi
fi

if [[ -n "$MKCERT" ]]; then
  yellow "Ensuring local CA is trusted (mkcert -install)…" >&2
  "$MKCERT" -install >/dev/null 2>&1 || yellow "(mkcert -install skipped or needs your password once.)" >&2
  green "Generating PEM for ${LOCAL_HOST}, localhost, 127.0.0.1, ::1 → $SSL_DIR" >&2
  if "$MKCERT" -cert-file "$CERT" -key-file "$KEY" "$LOCAL_HOST" localhost 127.0.0.1 ::1; then
    green "mkcert files updated (browser hostname should match one of these names)." >&2
  else
    yellow "mkcert -cert-file failed; see https://github.com/FiloSottile/mkcert#installation" >&2
  fi
else
  red "Could not install mkcert automatically (no apt/dnf/pacman/zypper package or Homebrew)." >&2
  yellow "Install manually: https://github.com/FiloSottile/mkcert#installation then re-run: ./scripts/nextpresskit setup" >&2
fi

_esc_host_dots() { printf '%s' "$1" | sed 's/\./\\./g'; }
if [[ -r /etc/hosts ]] && ! grep -qE "^[0-9.]+\s+$(_esc_host_dots "$LOCAL_HOST")(\s|$)" /etc/hosts 2>/dev/null; then
  yellow "Add this line to /etc/hosts if you use https://${LOCAL_HOST} :" >&2
  yellow "  127.0.0.1    ${LOCAL_HOST}" >&2
fi

os="$(uname -s 2>/dev/null || echo unknown)"
if [[ "$os" == Linux ]] && command -v nginx >/dev/null 2>&1; then
  if [[ -f "$CERT" && -f "$KEY" ]]; then
    green "Installing Nginx site (HTTPS, default PEM paths)…" >&2
    if bash "$ROOT/scripts/deploy" apply-nginx --no-tls-menu; then
      green "Nginx updated. Try: https://${LOCAL_HOST} (and/or https://localhost)" >&2
    else
      yellow "Nginx step failed (fix: sudo nginx -t). Then: ./scripts/nextpresskit deploy-apply-nginx --no-tls-menu" >&2
    fi
  else
    yellow "No PEM files yet — skipping Nginx HTTPS. After mkcert works, run:" >&2
    yellow "  ./scripts/nextpresskit deploy-apply-nginx --no-tls-menu" >&2
  fi
elif [[ "$os" == Linux ]]; then
  yellow "nginx not installed — PEM files (if any) are under $SSL_DIR" >&2
elif [[ "$os" == Darwin ]]; then
  yellow "macOS: PEM dir $SSL_DIR — use brew nginx or ./scripts/deploy apply-nginx (see deploy/generated/)." >&2
fi

exit 0

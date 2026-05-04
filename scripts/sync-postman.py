#!/usr/bin/env python3
"""Rewrite Postman environment JSON from .env.example + .env + process env.

Repo templates live in postman-templates/ (tracked). A gitignored postman/ directory
is created on demand: any missing *.json (and README.md) is copied from templates first.

Collections under postman/ only use {{base_url}} and env vars; they are not modified.

Tier base URLs (first match wins):
  - POSTMAN_LOCAL_BASE_URL, POSTMAN_DEV_BASE_URL, POSTMAN_STAGING_BASE_URL, POSTMAN_PRODUCTION_BASE_URL
  - or NEXTPRESS_PUBLIC_HOST for Local only → https://<host> when set
  - else documented placeholders (nextpresskit.local, api-*.example.com)

Shared values from env files (with shell overrides for these keys):
  SEED_SUPERADMIN_EMAIL, SEED_SUPERADMIN_PASSWORD, JWT_AUTH_SOURCE

Token variables (access_token, refresh_token, admin_access_token) are left unchanged
unless POSTMAN_CLEAR_TOKENS=1 (then set to empty string).
"""
from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
POSTMAN_DIR = ROOT / "postman"
TEMPLATE_DIR = ROOT / "postman-templates"

ENV_KEY_RE = re.compile(r"^([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)$")

DEFAULT_BASE = {
    "NextPressKit-Local.postman_environment.json": "https://nextpresskit.local",
    "NextPressKit-Dev.postman_environment.json": "https://api-dev.example.com",
    "NextPressKit-Staging.postman_environment.json": "https://api-staging.example.com",
    "NextPressKit-Production.postman_environment.json": "https://api.example.com",
}

SHARED_KEYS_FROM_DOTENV = (
    "SEED_SUPERADMIN_EMAIL",
    "SEED_SUPERADMIN_PASSWORD",
    "JWT_AUTH_SOURCE",
)


def parse_env_file(path: Path) -> dict[str, str]:
    out: dict[str, str] = {}
    if not path.is_file():
        return out
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        m = ENV_KEY_RE.match(line)
        if not m:
            continue
        key, val = m.group(1), m.group(2).strip()
        if val.startswith('"') and val.endswith('"') and len(val) >= 2:
            val = val[1:-1].replace(r"\"", '"')
        elif val.startswith("'") and val.endswith("'") and len(val) >= 2:
            val = val[1:-1]
        out[key] = val
    return out


def merged_config() -> dict[str, str]:
    """.env.example < .env < os.environ (for known keys and POSTMAN_*)."""
    cfg = parse_env_file(ROOT / ".env.example")
    cfg.update(parse_env_file(ROOT / ".env"))
    # Shell wins for dotenv keys we care about and all POSTMAN_* / NEXTPRESS_PUBLIC_HOST
    for k in SHARED_KEYS_FROM_DOTENV:
        v = os.environ.get(k)
        if v is not None and v != "":
            cfg[k] = v
    nph = os.environ.get("NEXTPRESS_PUBLIC_HOST")
    if nph is not None and nph.strip() != "":
        cfg["NEXTPRESS_PUBLIC_HOST"] = nph.strip()
    for k, v in os.environ.items():
        if k.startswith("POSTMAN_") and v != "":
            cfg[k] = v
    return cfg


def local_base_url(cfg: dict[str, str]) -> str:
    v = cfg.get("POSTMAN_LOCAL_BASE_URL", "").strip()
    if v:
        return v
    host = cfg.get("NEXTPRESS_PUBLIC_HOST", "").strip()
    if host:
        return f"https://{host}"
    return DEFAULT_BASE["NextPressKit-Local.postman_environment.json"]


def base_for_file(name: str, cfg: dict[str, str]) -> str:
    if name == "NextPressKit-Local.postman_environment.json":
        return local_base_url(cfg)
    overrides = {
        "NextPressKit-Dev.postman_environment.json": "POSTMAN_DEV_BASE_URL",
        "NextPressKit-Staging.postman_environment.json": "POSTMAN_STAGING_BASE_URL",
        "NextPressKit-Production.postman_environment.json": "POSTMAN_PRODUCTION_BASE_URL",
    }
    key = overrides.get(name)
    if key:
        v = cfg.get(key, "").strip()
        if v:
            return v
    return DEFAULT_BASE.get(name, "https://localhost")


def sync_file(path: Path, cfg: dict[str, str], *, dry_run: bool, clear_tokens: bool) -> list[str]:
    name = path.name
    data = json.loads(path.read_text(encoding="utf-8"))
    base = base_for_file(name, cfg)
    email = cfg.get("SEED_SUPERADMIN_EMAIL", "superadmin@nextpresskit.local")
    password = cfg.get("SEED_SUPERADMIN_PASSWORD", "SuperAdmin123!")
    jwt_src = (cfg.get("JWT_AUTH_SOURCE") or "cookie").strip() or "cookie"

    changes: list[str] = []
    for entry in data.get("values", []):
        key = entry.get("key")
        if key == "base_url" and entry.get("value") != base:
            changes.append(f"  {name}: base_url → {base!r}")
            entry["value"] = base
        elif key == "superadmin_email" and entry.get("value") != email:
            changes.append(f"  {name}: superadmin_email → {email!r}")
            entry["value"] = email
        elif key == "superadmin_password" and entry.get("value") != password:
            changes.append(f"  {name}: superadmin_password → ***")
            entry["value"] = password
        elif key == "jwt_auth_source" and entry.get("value") != jwt_src:
            changes.append(f"  {name}: jwt_auth_source → {jwt_src!r}")
            entry["value"] = jwt_src
        elif clear_tokens and key in ("access_token", "refresh_token", "admin_access_token"):
            if entry.get("value") != "":
                changes.append(f"  {name}: {key} cleared")
            entry["value"] = ""

    if changes and not dry_run:
        path.write_text(json.dumps(data, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    return changes


def ensure_postman_from_templates() -> int:
    """Copy missing JSON and README from postman-templates/ into gitignored postman/."""
    if not TEMPLATE_DIR.is_dir():
        print(f"Missing directory: {TEMPLATE_DIR}", file=sys.stderr)
        return 1
    templates = list(TEMPLATE_DIR.glob("*.json"))
    if not templates:
        print(f"No *.json in {TEMPLATE_DIR}", file=sys.stderr)
        return 1
    POSTMAN_DIR.mkdir(parents=True, exist_ok=True)
    copied = 0
    for src in sorted(templates):
        dest = POSTMAN_DIR / src.name
        if not dest.exists():
            shutil.copy2(src, dest)
            copied += 1
    readme_src = TEMPLATE_DIR / "README.md"
    readme_dest = POSTMAN_DIR / "README.md"
    if readme_src.is_file() and not readme_dest.exists():
        shutil.copy2(readme_src, readme_dest)
        copied += 1
    if copied:
        print(f"Seeded postman/ from postman-templates/ ({copied} new file(s)).")
    return 0


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.split("\n\n")[0])
    ap.add_argument(
        "--dry-run",
        action="store_true",
        help="Print changes only; do not write files",
    )
    args = ap.parse_args()
    clear_tokens = os.environ.get("POSTMAN_CLEAR_TOKENS", "").strip() in ("1", "true", "yes")

    os.chdir(ROOT)
    rc = ensure_postman_from_templates()
    if rc != 0:
        return rc
    cfg = merged_config()

    env_files = sorted(POSTMAN_DIR.glob("*.postman_environment.json"))
    if not env_files:
        print("No Postman environment files in postman/ after seeding.", file=sys.stderr)
        return 1

    all_changes: list[str] = []
    for p in env_files:
        all_changes.extend(sync_file(p, cfg, dry_run=args.dry_run, clear_tokens=clear_tokens))

    collections = list(POSTMAN_DIR.glob("*.postman_collection.json"))
    print(
        f"Postman: {len(env_files)} environment file(s); "
        f"{len(collections)} collection(s) unchanged (URLs use {{{{base_url}}}})."
    )
    if not all_changes:
        print("Already up to date.")
        return 0
    print("Updates:" if not args.dry_run else "Would update:")
    print("\n".join(all_changes))
    if args.dry_run:
        print("\nRe-run without --dry-run to write.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())

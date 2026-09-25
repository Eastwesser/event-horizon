#!/usr/bin/env python3
"""Dev-only idempotent user seed (bcrypt + psql).

Roles: user | author | admin

  cp scripts/.env.seed.admin.example scripts/.env.seed.admin
  make seed-admin

  # Author / user for three-role QA:
  make seed-author
  make seed-user

Env (overrides file):
  SEED_ROLE, SEED_EMAIL, SEED_PASSWORD, SEED_LAMPS, SEED_TICKETS
  SEED_GRANT_SUB=1|0  (default: 1 for admin, 0 for user/author)
  SEED_ADMIN_* aliases still work for admin.
"""
from __future__ import annotations

import os
import subprocess
import sys
from pathlib import Path
from urllib.parse import urlparse, unquote

try:
    import bcrypt
except ImportError:
    print("seed: need bcrypt — pip install bcrypt", file=sys.stderr)
    sys.exit(1)

ROOT = Path(__file__).resolve().parents[1]
ALLOWED_ROLES = frozenset({"user", "author", "admin"})


def load_dotenv(path: Path) -> None:
    if not path.is_file():
        return
    for raw in path.read_text().splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, val = line.partition("=")
        key, val = key.strip(), val.strip().strip("'\"")
        os.environ.setdefault(key, val)


def getenv(key: str, default: str = "") -> str:
    return os.environ.get(key, default).strip()


def dsn_to_psql_env(dsn: str) -> dict[str, str]:
    u = urlparse(dsn)
    env = os.environ.copy()
    env["PGHOST"] = u.hostname or "localhost"
    env["PGPORT"] = str(u.port or 5432)
    env["PGUSER"] = unquote(u.username or "eventhorizon")
    env["PGPASSWORD"] = unquote(u.password or "eventhorizon")
    env["PGDATABASE"] = (u.path or "/eventhorizon").lstrip("/") or "eventhorizon"
    return env


def psql(dsn: str, sql: str) -> str:
    env = dsn_to_psql_env(dsn)
    proc = subprocess.run(
        ["psql", "-v", "ON_ERROR_STOP=1", "-Atq", "-c", sql],
        env=env,
        capture_output=True,
        text=True,
    )
    if proc.returncode != 0:
        print(proc.stderr or proc.stdout, file=sys.stderr)
        sys.exit(proc.returncode)
    for line in proc.stdout.splitlines():
        line = line.strip()
        if line and not line.upper().startswith("INSERT"):
            return line
    return ""


def main() -> None:
    role = getenv("SEED_ROLE", "admin").lower()
    if role not in ALLOWED_ROLES:
        print(f"seed: SEED_ROLE must be one of {sorted(ALLOWED_ROLES)}", file=sys.stderr)
        sys.exit(1)

    # Prefer role-specific env file, fall back to admin file for shared DSNs.
    role_env = ROOT / f"scripts/.env.seed.{role}"
    admin_env = ROOT / "scripts" / ".env.seed.admin"
    load_dotenv(admin_env)
    if role_env.is_file() and role_env != admin_env:
        # Role file wins for unset keys — reload after clear would be complex;
        # instead: load role file with forced overwrite for SEED_* keys.
        for raw in role_env.read_text().splitlines():
            line = raw.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, _, val = line.partition("=")
            key, val = key.strip(), val.strip().strip("'\"")
            if key.startswith("SEED_") or key in ("ALLOW_DEV_SEED", "APP_ENV"):
                os.environ[key] = val

    if getenv("APP_ENV", "dev") == "production" and getenv("ALLOW_DEV_SEED") != "1":
        print("seed: refusing production without ALLOW_DEV_SEED=1", file=sys.stderr)
        sys.exit(1)

    defaults = {
        "admin": ("admin@eventhorizon.local", "1000000", "1000000", "1"),
        "author": ("author@eventhorizon.local", "5000", "5000", "1"),
        "user": ("player@eventhorizon.local", "1000", "1000", "0"),
    }
    def_email, def_lamps, def_tickets, def_sub = defaults[role]

    email = (
        getenv("SEED_EMAIL")
        or getenv("SEED_ADMIN_EMAIL")
        or def_email
    )
    password = getenv("SEED_PASSWORD") or getenv("SEED_ADMIN_PASSWORD")
    if not password:
        # Role-specific default passwords (dev only)
        password = {
            "admin": "",
            "author": "changeme-dev-author",
            "user": "changeme-dev-user",
        }.get(role, "")
    if not password:
        print(
            "seed: password empty — set SEED_PASSWORD or SEED_ADMIN_PASSWORD "
            f"(or create {role_env.name})",
            file=sys.stderr,
        )
        if role == "admin" and not admin_env.is_file():
            print(
                f"  cp scripts/.env.seed.admin.example scripts/.env.seed.admin",
                file=sys.stderr,
            )
        sys.exit(1)

    lamps = getenv("SEED_LAMPS") or getenv("SEED_ADMIN_LAMPS") or def_lamps
    tickets = getenv("SEED_TICKETS") or getenv("SEED_ADMIN_TICKETS") or def_tickets
    grant_sub = getenv("SEED_GRANT_SUB") or def_sub
    grant_sub = grant_sub in ("1", "true", "yes", "on")

    auth_dsn = getenv(
        "SEED_AUTH_DSN",
        "postgres://eventhorizon:eventhorizon@localhost:5460/eventhorizon?sslmode=disable",
    )
    billing_dsn = getenv(
        "SEED_BILLING_DSN",
        "postgres://eventhorizon:eventhorizon@localhost:5462/eventhorizon_billing?sslmode=disable",
    )
    payment_dsn = getenv(
        "SEED_PAYMENT_DSN",
        "postgres://eventhorizon:eventhorizon@localhost:5467/eventhorizon_payment?sslmode=disable",
    )
    sub_plan = getenv("SEED_ADMIN_SUB_PLAN") or getenv("SEED_SUB_PLAN") or "present"
    sub_days = int(getenv("SEED_ADMIN_SUB_DAYS") or getenv("SEED_SUB_DAYS") or "3650")
    sub_plan_sql = sub_plan.replace("'", "''")
    role_sql = role.replace("'", "''")

    pw_hash = bcrypt.hashpw(password.encode(), bcrypt.gensalt(rounds=12)).decode()
    email_sql = email.replace("'", "''")
    hash_sql = pw_hash.replace("'", "''")

    user_id = psql(
        auth_dsn,
        f"""
        INSERT INTO users (email, password_hash, role)
        VALUES ('{email_sql}', '{hash_sql}', '{role_sql}')
        ON CONFLICT (email) DO UPDATE SET
            password_hash = EXCLUDED.password_hash,
            role = EXCLUDED.role,
            updated_at = CURRENT_TIMESTAMP
        RETURNING id;
        """,
    )
    if not user_id:
        print("seed: no user id returned", file=sys.stderr)
        sys.exit(1)

    for currency, balance in (("lamps", lamps), ("tickets", tickets)):
        psql(
            billing_dsn,
            f"""
            INSERT INTO user_currencies (user_id, currency_type, balance, updated_at)
            VALUES ('{user_id}'::uuid, '{currency}', {int(balance)}, CURRENT_TIMESTAMP)
            ON CONFLICT (user_id, currency_type) DO UPDATE SET
                balance = EXCLUDED.balance,
                updated_at = CURRENT_TIMESTAMP;
            """,
        )

    sub_id = ""
    if grant_sub:
        psql(
            payment_dsn,
            f"""
            UPDATE subscriptions
            SET status = 'expired', updated_at = NOW()
            WHERE user_id = '{user_id}' AND status = 'active';
            """,
        )
        sub_id = psql(
            payment_dsn,
            f"""
            INSERT INTO subscriptions (
                id, user_id, plan, status, amount_rub, payment_id,
                starts_at, expires_at, created_at, updated_at
            ) VALUES (
                gen_random_uuid(),
                '{user_id}',
                '{sub_plan_sql}',
                'active',
                0,
                NULL,
                NOW(),
                NOW() + INTERVAL '{sub_days} days',
                NOW(),
                NOW()
            )
            RETURNING id;
            """,
        )

    print(f"seed: ok ({role})")
    print(f"  user_id:  {user_id}")
    print(f"  email:    {email}")
    print(f"  role:     {role}")
    print(f"  lamps:    {lamps}")
    print(f"  tickets:  {tickets}")
    if grant_sub:
        print(f"  sub_plan: {sub_plan} (active, {sub_days}d)")
        print(f"  sub_id:   {sub_id or '(none)'}")
    else:
        print("  sub:      (skipped)")
    print("  password: (from env / .env.seed.* — not printed)")
    print(f"Re-run: SEED_ROLE={role} make seed-dev   (idempotent upsert)")


if __name__ == "__main__":
    main()

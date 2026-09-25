# Seed an idempotent admin@… god account (1M lamps + tickets).
# Credentials live in scripts/.env.seed.admin (gitignored) — never the frontend.

## Setup (once)

```bash
cp scripts/.env.seed.admin.example scripts/.env.seed.admin
# edit SEED_ADMIN_PASSWORD
pip install bcrypt psycopg2-binary   # if needed
```

## Run

```bash
make seed-admin
```

Requires auth + billing Postgres from `docker compose` (ports 5460 / 5462).

## Login

Use `SEED_ADMIN_EMAIL` / `SEED_ADMIN_PASSWORD` from your local `.env.seed.admin`.

(Go reference implementation also lives under `scripts/seed-admin/` if you prefer `go run`.)

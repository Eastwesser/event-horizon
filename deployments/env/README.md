# Per-service env templates (local `go run`, not cluster compose)

Cluster compose reads **repo-root `.env`** (see [`.env.example`](../../.env.example)).

For running a single service locally:

```bash
cp .env.example .env                    # once at repo root
cp deployments/env/auth.env.template deployments/env/auth.env
# auth.env uses ${POSTGRES_USER} — source root .env first:
set -a && source .env && source deployments/env/auth.env && set +a
(cd services/auth && go run ./cmd/main.go)
```

Never commit `deployments/env/*.env` or repo-root `.env`.

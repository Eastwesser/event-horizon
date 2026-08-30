# GitGuardian — assessment (26.08.2026)

Commit **`e6bb5ab`** (`chore(devops): fix compose env/JWT loading and add Auth Patroni stubs`) triggered GitGuardian on **hardcoded database passwords** in Patroni stub files.

---

## Verdict: no crucial production leaks in that commit

| Finding | Risk | Action |
|---------|------|--------|
| `PATRONI_SUPERUSER_PASSWORD: patroni` | **Low** — documented dev stub | OK for stubs; use k8s Secrets in prod |
| `PATRONI_REPLICATION_PASSWORD: replicator` | **Low** | Same |
| `PATRONI_admin_PASSWORD: eventhorizon` | **Low** — matches compose dev DB user | Same class as `docker-compose.cluster.yml` postgres creds |
| `JWT_SECRET=${JWT_SECRET}` in compose | **None in git** — env substitution | Ensure `.env` exists locally, never commit |
| ClickHouse `default-network.xml` | **Config, not a secret** | Allows Docker-network access by design |

**You committed stubs, not live tokens.** GitGuardian is doing its job — it cannot distinguish “example postgres password” from a real one.

---

## What would be crucial (not seen in e6bb5ab)

These would require **immediate rotation** if they appeared in git history:

- Live **JWT_SECRET** or signing keys
- **Docker Hub** / **GitHub** personal access tokens (`dckr_pat_…`, `ghp_…`)
- **YooKassa / Boosty / Telegram** API keys
- Production **`.env`** with real credentials

Checks (26.08.2026):

- `.env` is in `.gitignore` and has **no git history** in this repo
- No `dckr_pat` / `ghp_` patterns in tracked source (excluding confluence history notes)

Past note ([`INSPECTOR_REPORT.md`](../13.08.2026/INSPECTOR_REPORT.md)): local `.env` once held Docker Hub + Ansible passwords **on disk only**. If that file was ever copied or shared, rotate those credentials regardless of GitGuardian.

---

## If alerts keep firing

1. **Do nothing critical** for stub passwords — they are intentional placeholders.
2. Optional: replace compose literals with `${PATRONI_SUPERUSER_PASSWORD}` + document in `.env.example` (never commit `.env`).
3. Optional: [GitGuardian allowlist](https://docs.gitguardian.com/internal-repositories-monitoring/integrations/gitguardian_yml) for `deployments/patroni/**` with a comment that values are dev stubs.
4. **Never** allowlist `.env` or real token patterns.

---

## Related

- Patroni roadmap: [`deployments/patroni/README.md`](../../../deployments/patroni/README.md)
- Deploy JWT note: [`DEPLOY_DIAGNOSIS.md`](../22.08.2026/DEPLOY_DIAGNOSIS.md)
